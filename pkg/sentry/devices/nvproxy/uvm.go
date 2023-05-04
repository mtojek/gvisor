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
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/abi/nvgpu"
	"gvisor.dev/gvisor/pkg/context"
	"gvisor.dev/gvisor/pkg/errors/linuxerr"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/sentry/arch"
	"gvisor.dev/gvisor/pkg/sentry/kernel"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
	"gvisor.dev/gvisor/pkg/usermem"
)

// uvmDevice implements vfs.Device for /dev/nvidia-uvm.
//
// +stateify savable
type uvmDevice struct {
	nvp *nvproxy
}

// Open implements vfs.Device.Open.
func (dev *uvmDevice) Open(ctx context.Context, mnt *vfs.Mount, vfsd *vfs.Dentry, opts vfs.OpenOptions) (*vfs.FileDescription, error) {
	hostFD, err := unix.Openat(-1, "/dev/nvidia-uvm", unix.O_RDWR|unix.O_NOFOLLOW, 0)
	if err != nil {
		ctx.Warningf("nvproxy: failed to open host /dev/nvidia-uvm: %v", err)
		return nil, err
	}
	fd := &uvmFD{
		nvp:    dev.nvp,
		hostFD: int32(hostFD),
	}
	if err := fd.vfsfd.Init(fd, opts.Flags, mnt, vfsd, &vfs.FileDescriptionOptions{
		UseDentryMetadata: true,
	}); err != nil {
		unix.Close(hostFD)
		return nil, err
	}
	return &fd.vfsfd, nil
}

// uvmFD implements vfs.FileDescriptionImpl for /dev/nvidia-uvm.
//
// uvmFD is not savable; we do not implement save/restore of host GPU state.
type uvmFD struct {
	vfsfd vfs.FileDescription
	vfs.FileDescriptionDefaultImpl
	vfs.DentryMetadataFileDescriptionImpl
	vfs.NoLockFD

	nvp    *nvproxy
	hostFD int32
}

// Release implements vfs.FileDescriptionImpl.Release.
func (fd *uvmFD) Release(context.Context) {
	unix.Close(int(fd.hostFD))
}

// Ioctl implements vfs.FileDescriptionImpl.Ioctl.
func (fd *uvmFD) Ioctl(ctx context.Context, uio usermem.IO, sysno uintptr, args arch.SyscallArguments) (uintptr, error) {
	cmd := args[1].Uint()
	argPtr := args[2].Pointer()

	t := kernel.TaskFromContext(ctx)
	if t == nil {
		panic("Ioctl should be called from a task context")
	}

	ui := uvmIoctlState{
		fd:              fd,
		ctx:             ctx,
		t:               t,
		cmd:             cmd,
		ioctlParamsAddr: argPtr,
	}

	switch cmd {
	case nvgpu.UVM_INITIALIZE:
		return uvmInitialize(&ui)

	case nvgpu.UVM_DEINITIALIZE:
		return uvmIoctlInvoke[byte](&ui, nil)

	case nvgpu.UVM_PAGEABLE_MEM_ACCESS:
		return uvmIoctlSimple[nvgpu.UVM_PAGEABLE_MEM_ACCESS_PARAMS](&ui)

	default:
		ctx.Warningf("nvproxy: unknown uvm ioctl %d", cmd)
		return 0, linuxerr.EINVAL
	}
}

// uvmIoctlState holds the state of a call to uvmFD.Ioctl().
type uvmIoctlState struct {
	fd              *uvmFD
	ctx             context.Context
	t               *kernel.Task
	cmd             uint32
	ioctlParamsAddr hostarch.Addr
}

func uvmIoctlSimple[Params any, PParams marshalPtr[Params]](ui *uvmIoctlState) (uintptr, error) {
	var ioctlParams Params
	if _, err := (PParams)(&ioctlParams).CopyIn(ui.t, ui.ioctlParamsAddr); err != nil {
		return 0, err
	}
	n, err := uvmIoctlInvoke(ui, &ioctlParams)
	if err != nil {
		return n, err
	}
	if _, err := (PParams)(&ioctlParams).CopyOut(ui.t, ui.ioctlParamsAddr); err != nil {
		return n, err
	}
	return n, nil
}

func uvmInitialize(ui *uvmIoctlState) (uintptr, error) {
	var ioctlParams nvgpu.UVM_INITIALIZE_PARAMS
	if _, err := ioctlParams.CopyIn(ui.t, ui.ioctlParamsAddr); err != nil {
		return 0, err
	}
	sentryIoctlParams := nvgpu.UVM_INITIALIZE_PARAMS{
		// This is necessary to share the host UVM FD between sentry and
		// application processes.
		Flags:    ioctlParams.Flags | nvgpu.UVM_INIT_FLAGS_MULTI_PROCESS_SHARING_MODE,
		RMStatus: ioctlParams.RMStatus,
	}
	n, err := uvmIoctlInvoke(ui, &sentryIoctlParams)
	if err != nil {
		return n, err
	}
	outIoctlParams := nvgpu.UVM_INITIALIZE_PARAMS{
		// Only expose the MULTI_PROCESS_SHARING_MODE flag if it was present in
		// ioctlParams.
		Flags:    sentryIoctlParams.Flags &^ (^ioctlParams.Flags & nvgpu.UVM_INIT_FLAGS_MULTI_PROCESS_SHARING_MODE),
		RMStatus: sentryIoctlParams.RMStatus,
	}
	if _, err := outIoctlParams.CopyOut(ui.t, ui.ioctlParamsAddr); err != nil {
		return n, err
	}
	return n, nil
}
