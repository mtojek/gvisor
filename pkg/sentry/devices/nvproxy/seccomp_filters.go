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
	"gvisor.dev/gvisor/pkg/abi/linux"
	"gvisor.dev/gvisor/pkg/abi/nvgpu"
	"gvisor.dev/gvisor/pkg/seccomp"
)

// Filters returns seccomp-bpf filters for this package.
func Filters() seccomp.SyscallRules {
	nonNegativeFD := seccomp.LessThanOrEqual(0x7fff_ffff /* max int32 */)
	notIocSizeMask := ^(((uintptr(1) << linux.IOC_SIZEBITS) - 1) << linux.IOC_SIZESHIFT) // for ioctls taking arbitrary size
	return seccomp.SyscallRules{
		unix.SYS_OPENAT: []seccomp.Rule{
			{
				// All paths that we openat() are absolute, so we pass a dirfd
				// of -1 (which is invalid for relative paths, but ignored for
				// absolute paths) to hedge against bugs involving AT_FDCWD or
				// real dirfds.
				seccomp.EqualTo(^uintptr(0)),
				seccomp.MatchAny{},
				seccomp.MaskedEqual(unix.O_NOFOLLOW|unix.O_CREAT, unix.O_NOFOLLOW),
				seccomp.MatchAny{},
			},
		},
		unix.SYS_IOCTL: []seccomp.Rule{
			{
				nonNegativeFD,
				seccomp.MaskedEqual(notIocSizeMask, frontendIoctlCmd(nvgpu.NV_ESC_CARD_INFO, 0)),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_CHECK_VERSION_STR, uint32(nvgpu.SizeofRMAPIVersion))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_REGISTER_FD, uint32(nvgpu.SizeofIoctlRegisterFD))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_SYS_PARAMS, uint32(nvgpu.SizeofIoctlSysParams))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_RM_FREE, uint32(nvgpu.SizeofNVOS00Parameters))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_RM_CONTROL, uint32(nvgpu.SizeofNVOS54Parameters))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_RM_ALLOC, uint32(nvgpu.SizeofNVOS21Parameters))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(frontendIoctlCmd(nvgpu.NV_ESC_RM_ALLOC, uint32(nvgpu.SizeofNVOS64Parameters))),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(nvgpu.UVM_INITIALIZE),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(nvgpu.UVM_DEINITIALIZE),
			},
			{
				nonNegativeFD,
				seccomp.EqualTo(nvgpu.UVM_PAGEABLE_MEM_ACCESS),
			},
		},
	}
}
