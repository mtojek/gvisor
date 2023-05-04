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

import (
	"unsafe"
)

// Frontend ioctl parameter struct sizes.
const (
	SizeofIoctlRegisterFD  = unsafe.Sizeof(IoctlRegisterFD{})
	SizeofRMAPIVersion     = unsafe.Sizeof(RMAPIVersion{})
	SizeofIoctlSysParams   = unsafe.Sizeof(IoctlSysParams{})
	SizeofNVOS00Parameters = unsafe.Sizeof(NVOS00Parameters{})
	SizeofNVOS21Parameters = unsafe.Sizeof(NVOS21Parameters{})
	SizeofNVOS54Parameters = unsafe.Sizeof(NVOS54Parameters{})
	SizeofNVOS64Parameters = unsafe.Sizeof(NVOS64Parameters{})
)
