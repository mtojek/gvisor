// Copyright 2023 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nvproxy

import (
	"fmt"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/abi/nvgpu"
	"gvisor.dev/gvisor/pkg/context"
	"gvisor.dev/gvisor/pkg/errors/linuxerr"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"gvisor.dev/gvisor/pkg/usermem"
)

// frontendDevice implements vfs.Device for /dev/nvidia# and /dev/nvidiactl.
//
// +stateify savable
type frontendDevice struct {
	nvp   *nvproxy
	minor uint32
}

// Open implements vfs.Device.Open.
func (dev *frontendDevice) Open(ctx context.Context, mnt *vfs.Mount, vfsd *vfs.Dentry, opts vfs.OpenOptions) (*vfs.FileDescription, error) {
	var hostPath string
	if dev.minor == nvgpu.NV_CONTROL_DEVICE_MINOR {
		hostPath = "/dev/nvidiactl"
	} else {
		hostPath = fmt.Sprintf("/dev/nvidia%d", dev.minor)
	}
	hostFD, err := unix.Openat(-1, hostPath, unix.O_RDONLY|unix.O_NOFOLLOW, 0)
	if err != nil {
		ctx.Warningf("nvproxy: failed to open host %s: %v", hostPath, err)
		return nil, err
	}
	fd := &frontendFD{
		nvp:       dev.nvp,
		hostFD:    int32(hostFD),
		isControl: dev.minor == nvgpu.NV_CONTROL_DEVICE_MINOR,
	}
	if err := fd.vfsfd.Init(fd, opts.Flags, mnt, vfsd, &vfs.FileDescriptionOptions{
		UseDentryMetadata: true,
	}); err != nil {
		unix.Close(hostFD)
		return nil, err
	}
	return &fd.vfsfd, nil
}

// frontendFD implements vfs.FileDescriptionImpl for /dev/nvidia# and
// /dev/nvidiactl.
//
// frontendFD is not savable; we do not implement save/restore of host GPU
// state.
type frontendFD struct {
	vfsfd vfs.FileDescription
	vfs.FileDescriptionDefaultImpl
	vfs.DentryMetadataFileDescriptionImpl
	vfs.NoLockFD

	nvp       *nvproxy
	hostFD    int32
	isControl bool
}

// Release implements vfs.FileDescriptionImpl.Release.
func (fd *frontendFD) Release(context.Context) {
	unix.Close(int(fd.hostFD))
}

// Ioctl implements vfs.FileDescriptionImpl.Ioctl.
func (fd *frontendFD) Ioctl(ctx context.Context, uio usermem.IO, sysno uintptr, args arch.SyscallArguments) (uintptr, error) {
	cmd := args[1].Uint()
	nr := linux.IOC_NR(cmd)
	argPtr := args[2].Pointer()
	argSize := linux.IOC_SIZE(cmd)

	t := kernel.TaskFromContext(ctx)
	if t == nil {
		panic("Ioctl should be called from a task context")
	}

	fi := frontendIoctlState{
		fd:              fd,
		ctx:             ctx,
		t:               t,
		nr:              nr,
		ioctlParamsAddr: argPtr,
		ioctlParamsSize: argSize,
	}

	// nr determines the argument type.
	// See src/nvidia/arch/nvalloc/unix/src/escape.c:RmIoctl() for NV_ESC_RM_*
	// ioctls and NV_ESC_REGISTER_FD, and
	// kernel-open/nvidia/nv.c:nvidia_ioctl() for others.
	switch nr {
	case
		nvgpu.NV_ESC_CARD_INFO,         // nv_ioctl_card_info_t
		nvgpu.NV_ESC_CHECK_VERSION_STR, // nv_rm_api_version_t
		nvgpu.NV_ESC_SYS_PARAMS,        // nv_ioctl_sys_params_t
		nvgpu.NV_ESC_RM_FREE:           // NVOS00_PARAMETERS
		return frontendIoctlSimple(&fi)

	case nvgpu.NV_ESC_REGISTER_FD:
		return frontendRegisterFD(&fi)

	case nvgpu.NV_ESC_NUMA_INFO:
		// Rejecting this is non-fatal. Figure out how to proxy it in the
		// future.
		ctx.Infof("nvproxy: rejecting NV_ESC_NUMA_INFO")
		return 0, linuxerr.EINVAL

	case nvgpu.NV_ESC_RM_CONTROL:
		return rmControl(&fi)

	case nvgpu.NV_ESC_RM_ALLOC:
		return rmAlloc(&fi)

	default:
		ctx.Warningf("nvproxy: unknown frontend ioctl %d == %#x (argSize=%d, cmd=%#x)", nr, nr, argSize, cmd)
		return 0, linuxerr.EINVAL
	}
}

func frontendIoctlCmd(nr, argSize uint32) uintptr {
	return uintptr(linux.IOWR(nvgpu.NV_IOCTL_MAGIC, nr, argSize))
}

// frontendIoctlState holds the state of a call to frontendFD.Ioctl().
type frontendIoctlState struct {
	fd              *frontendFD
	ctx             context.Context
	t               *kernel.Task
	nr              uint32
	ioctlParamsAddr hostarch.Addr
	ioctlParamsSize uint32
}

// frontendIoctlSimple implements a frontend ioctl whose parameters don't
// contain any pointers or filtered fields and consequently don't need to be
// typed by the sentry.
func frontendIoctlSimple(fi *frontendIoctlState) (uintptr, error) {
	if fi.ioctlParamsSize == 0 {
		return frontendIoctlInvoke[byte](fi, nil)
	}

	ioctlParams := make([]byte, fi.ioctlParamsSize)
	if _, err := fi.t.CopyInBytes(fi.ioctlParamsAddr, ioctlParams); err != nil {
		return 0, err
	}
	n, err := frontendIoctlInvoke(fi, &ioctlParams[0])
	if err != nil {
		return n, err
	}
	if _, err := fi.t.CopyOutBytes(fi.ioctlParamsAddr, ioctlParams); err != nil {
		return n, err
	}
	return n, nil
}

func frontendRegisterFD(fi *frontendIoctlState) (uintptr, error) {
	var ioctlParams nvgpu.IoctlRegisterFD
	if uintptr(fi.ioctlParamsSize) != nvgpu.SizeofIoctlRegisterFD {
		return 0, linuxerr.EINVAL
	}
	if _, err := ioctlParams.CopyIn(fi.t, fi.ioctlParamsAddr); err != nil {
		return 0, err
	}

	ctlFileGeneric, _ := fi.t.FDTable().Get(ioctlParams.CtlFD)
	if ctlFileGeneric == nil {
		return 0, linuxerr.EINVAL
	}
	defer ctlFileGeneric.DecRef(fi.ctx)
	ctlFile, ok := ctlFileGeneric.Impl().(*frontendFD)
	if !ok {
		return 0, linuxerr.EINVAL
	}

	sentryIoctlParams := nvgpu.IoctlRegisterFD{
		CtlFD: ctlFile.hostFD,
	}
	// The returned ctl_fd can't change, so skip copying out.
	return frontendIoctlInvoke(fi, &sentryIoctlParams)
}

func rmControl(fi *frontendIoctlState) (uintptr, error) {
	var ioctlParams nvgpu.NVOS54Parameters
	if uintptr(fi.ioctlParamsSize) != nvgpu.SizeofNVOS54Parameters {
		return 0, linuxerr.EINVAL
	}
	if _, err := ioctlParams.CopyIn(fi.t, fi.ioctlParamsAddr); err != nil {
		return 0, err
	}

	// Cmd determines the type of Params.
	if ioctlParams.Cmd&0x00008000 != 0 {
		// This is a "legacy GSS control" that is implemented by the GPU System
		// Processor (GSP). Conseqeuently, its parameters cannot reasonably
		// contain application pointers, and the control is in any case
		// undocumented.
		// See
		// src/nvidia/src/kernel/rmapi/entry_points.c:_nv04ControlWithSecInfo()
		// =>
		// src/nvidia/interface/deprecated/rmapi_deprecated_control.c:RmDeprecatedGetControlHandler()
		// =>
		// src/nvidia/interface/deprecated/rmapi_gss_legacy_control.c:RmGssLegacyRpcCmd().
		return rmControlSimple(fi, &ioctlParams)
	}
	// The type name is always `Cmd ~ s/CTRL_CMD/CTRL/` + "_PARAMS".
	switch ioctlParams.Cmd {
	case
		nvgpu.NV0000_CTRL_CMD_CLIENT_SET_INHERITED_SHARE_POLICY,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_ATTACHED_IDS,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_ID_INFO,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_ID_INFO_V2,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_PROBED_IDS,
		nvgpu.NV0000_CTRL_CMD_GPU_ATTACH_IDS,
		nvgpu.NV0000_CTRL_CMD_GPU_DETACH_IDS,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_PCI_INFO,
		nvgpu.NV0000_CTRL_CMD_GPU_QUERY_DRAIN_STATE,
		nvgpu.NV0000_CTRL_CMD_GPU_GET_MEMOP_ENABLE,
		nvgpu.NV0000_CTRL_CMD_SYNC_GPU_BOOST_GROUP_INFO,
		nvgpu.NV0080_CTRL_CMD_FB_GET_CAPS_V2,
		nvgpu.NV0080_CTRL_CMD_GPU_GET_NUM_SUBDEVICES,
		nvgpu.NV0080_CTRL_CMD_GPU_QUERY_SW_STATE_PERSISTENCE,
		nvgpu.NV0080_CTRL_CMD_GPU_GET_VIRTUALIZATION_MODE,
		nvgpu.NV0080_CTRL_CMD_GPU_GET_CLASSLIST_V2,
		nvgpu.NV0080_CTRL_CMD_HOST_GET_CAPS_V2,
		nvgpu.NV2080_CTRL_CMD_BUS_GET_PCI_INFO,
		nvgpu.NV2080_CTRL_CMD_BUS_GET_PCI_BAR_INFO,
		nvgpu.NV2080_CTRL_CMD_BUS_GET_INFO_V2,
		nvgpu.NV2080_CTRL_CMD_BUS_GET_PCIE_SUPPORTED_GPU_ATOMICS,
		nvgpu.NV2080_CTRL_CMD_CE_GET_ALL_CAPS,
		nvgpu.NV2080_CTRL_CMD_FB_GET_INFO_V2,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_INFO_V2,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_NAME_STRING,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_SIMULATION_INFO,
		nvgpu.NV2080_CTRL_CMD_GPU_QUERY_ECC_STATUS,
		nvgpu.NV2080_CTRL_CMD_GPU_QUERY_COMPUTE_MODE_RULES,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_GID_INFO,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_ENGINES_V2,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_ACTIVE_PARTITION_IDS,
		nvgpu.NV2080_CTRL_CMD_GPU_GET_COMPUTE_POLICY_CONFIG,
		nvgpu.NV2080_CTRL_CMD_GR_GET_GLOBAL_SM_ORDER,
		nvgpu.NV2080_CTRL_CMD_GR_GET_CAPS_V2,
		nvgpu.NV2080_CTRL_CMD_GR_GET_GPC_MASK,
		nvgpu.NV2080_CTRL_CMD_GR_GET_TPC_MASK,
		nvgpu.NV2080_CTRL_CMD_MC_GET_ARCH_INFO,
		nvgpu.NV2080_CTRL_CMD_TIMER_GET_GPU_CPU_TIME_CORRELATION_INFO:
		return rmControlSimple(fi, &ioctlParams)

	case nvgpu.NV0000_CTRL_CMD_SYSTEM_GET_BUILD_VERSION:
		return ctrlClientSystemGetBuildVersion(fi, &ioctlParams)

	case nvgpu.NV2080_CTRL_CMD_GR_GET_INFO:
		return ctrlSubdevGRGetInfo(fi, &ioctlParams)

	default:
		fi.ctx.Warningf("nvproxy: unknown control command %#x", ioctlParams.Cmd)
		return 0, linuxerr.EINVAL
	}
}

func rmAlloc(fi *frontendIoctlState) (uintptr, error) {
	// Copy in parameters and convert to NVOS64Parameters.
	var (
		ioctlParams nvgpu.NVOS64Parameters
		isNVOS64    bool
	)
	switch uintptr(fi.ioctlParamsSize) {
	case nvgpu.SizeofNVOS21Parameters:
		var buf nvgpu.NVOS21Parameters
		if _, err := buf.CopyIn(fi.t, fi.ioctlParamsAddr); err != nil {
			return 0, err
		}
		ioctlParams = nvgpu.NVOS64Parameters{
			HRoot:         buf.HRoot,
			HObjectParent: buf.HObjectParent,
			HObjectNew:    buf.HObjectNew,
			HClass:        buf.HClass,
			PAllocParms:   buf.PAllocParms,
			Status:        buf.Status,
		}
	case nvgpu.SizeofNVOS64Parameters:
		if _, err := ioctlParams.CopyIn(fi.t, fi.ioctlParamsAddr); err != nil {
			return 0, err
		}
		isNVOS64 = true
	default:
		return 0, linuxerr.EINVAL
	}

	// hClass determines the type of pAllocParms.
	// See src/nvidia/generated/g_allclasses.h for class constants.
	// See src/nvidia/src/kernel/rmapi/resource_list.h for table mapping class
	// ("External Class") to the type of pAllocParms ("Alloc Param Info") and
	// the class whose constructor interprets it ("Internal Class").
	switch ioctlParams.HClass {
	case nvgpu.NV01_ROOT, nvgpu.NV01_ROOT_NON_PRIV, nvgpu.NV01_ROOT_CLIENT:
		return rmAllocSimple[nvgpu.Handle](fi, &ioctlParams, isNVOS64)

	case nvgpu.NV01_DEVICE_0:
		return rmAllocSimple[nvgpu.NV0080_ALLOC_PARAMETERS](fi, &ioctlParams, isNVOS64)

	case nvgpu.NV20_SUBDEVICE_0:
		return rmAllocSimple[nvgpu.NV2080_ALLOC_PARAMETERS](fi, &ioctlParams, isNVOS64)

	default:
		fi.ctx.Warningf("nvproxy: unknown allocation class %#08x", ioctlParams.HClass)
		return 0, linuxerr.EINVAL
	}
}

// Unlike frontendIoctlSimple and rmControlSimple, rmAllocSimple requires the
// parameter type since the parameter's size is otherwise unknown.
func rmAllocSimple[Params any, PParams marshalPtr[Params]](fi *frontendIoctlState, ioctlParams *nvgpu.NVOS64Parameters, isNVOS64 bool) (uintptr, error) {
	if ioctlParams.PAllocParms.IsNull() {
		return rmAllocInvoke[byte](fi, ioctlParams, nil, isNVOS64)
	}

	var allocParams Params
	if _, err := (PParams)(&allocParams).CopyIn(fi.t, addrFromP64(ioctlParams.PAllocParms)); err != nil {
		return 0, err
	}
	n, err := rmAllocInvoke(fi, ioctlParams, &allocParams, isNVOS64)
	if err != nil {
		return n, err
	}
	if _, err := (PParams)(&allocParams).CopyOut(fi.t, addrFromP64(ioctlParams.PAllocParms)); err != nil {
		return n, err
	}
	return n, nil
}
