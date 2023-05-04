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

// UVM ioctl commands.
const (
	// From kernel-open/nvidia-uvm/uvm_linux_ioctl.h:
	UVM_INITIALIZE   = 0x30000001
	UVM_DEINITIALIZE = 0x30000002

	// From kernel-open/nvidia-uvm/uvm_ioctl.h:
	UVM_PAGEABLE_MEM_ACCESS = 39
)

// +marshal
type UVM_INITIALIZE_PARAMS struct {
	Flags    uint64
	RMStatus uint32
}

// UVM_INITIALIZE_PARAMS flags, from kernel-open/nvidia-uvm/uvm_types.h.
const (
	UVM_INIT_FLAGS_MULTI_PROCESS_SHARING_MODE = 0x2
)

// +marshal
type UVM_PAGEABLE_MEM_ACCESS_PARAMS struct {
	PageableMemAccess uint8
	Pad               [3]byte
	RMStatus          uint32
}
