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

package nvgpu

// NV_IOCTL_MAGIC is the "canonical" IOC_TYPE for frontend ioctls.
// The driver ignores IOC_TYPE, allowing any value to be passed.
const NV_IOCTL_MAGIC = uint32('F')

// Frontend ioctl numbers.
// Note that these are only the IOC_NR part of the ioctl command.
const (
	// From kernel-open/common/inc/nv-ioctl-numbers.h:
	NV_IOCTL_BASE            = 200
	NV_ESC_CARD_INFO         = NV_IOCTL_BASE + 0
	NV_ESC_REGISTER_FD       = NV_IOCTL_BASE + 1
	NV_ESC_CHECK_VERSION_STR = NV_IOCTL_BASE + 10
	NV_ESC_SYS_PARAMS        = NV_IOCTL_BASE + 14

	// From kernel-open/common/inc/nv-ioctl-numa.h:
	NV_ESC_NUMA_INFO = NV_IOCTL_BASE + 15

	// From src/nvidia/arch/nvalloc/unix/include/nv_escape.h:
	NV_ESC_RM_FREE    = 0x29
	NV_ESC_RM_CONTROL = 0x2a
	NV_ESC_RM_ALLOC   = 0x2b
)

// Frontend ioctl parameter structs.
// NV_ESC_RM_* ioctl parameter structs are from
// src/common/sdk/nvidia/inc/nvos.h.
// Other ioctl parameter structs are from kernel-open/common/inc/nv-ioctl.h.

// IoctlRegisterFD is nv_ioctl_register_fd_t, the parameter type for
// NV_ESC_REGISTER_FD.
//
// +marshal
type IoctlRegisterFD struct {
	CtlFD int32
}

// RMAPIVersion is nv_rm_api_version_t, the parameter type for
// NV_ESC_CHECK_VERSION_STR.
//
// +marshal
type RMAPIVersion struct {
	Cmd           uint32
	Reply         uint32
	VersionString [64]byte
}

// IoctlSysParams is nv_ioctl_sys_params_t, the parameter type for
// NV_ESC_SYS_PARAMS.
//
// +marshal
type IoctlSysParams struct {
	MemblockSize uint64
}

// NVOS00Parameters is NVOS00_PARAMETERS, the parameter type for
// NV_ESC_RM_FREE.
//
// +marshal
type NVOS00Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectOld    Handle
	Status        uint32
}

// NVOS21Parameters is NVOS21_PARAMETERS, one possible parameter type for
// NV_ESC_RM_ALLOC.
//
// +marshal
type NVOS21Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectNew    Handle
	HClass        uint32
	PAllocParms   P64
	Status        uint32
}

// NVOS54Parameters is NVOS54_PARAMETERS, the parameter type for
// NV_ESC_RM_CONTROL.
//
// +marshal
type NVOS54Parameters struct {
	HClient    Handle
	HObject    Handle
	Cmd        uint32
	Flags      uint32
	Params     P64
	ParamsSize uint32
	Status     uint32
}

// NVOS64Parameters is NVOS64_PARAMETERS, one possible parameter type for
// NV_ESC_RM_ALLOC.
//
// +marshal
type NVOS64Parameters struct {
	HRoot            Handle
	HObjectParent    Handle
	HObjectNew       Handle
	HClass           uint32
	PAllocParms      P64
	PRightsRequested P64
	Flags            uint32
	Status           uint32
}
