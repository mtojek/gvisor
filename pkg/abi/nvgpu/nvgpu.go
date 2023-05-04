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

// Package nvgpu tracks the ABI of the Nvidia GPU Linux kernel driver:
// https://github.com/NVIDIA/open-gpu-kernel-modules
package nvgpu

// Device numbers.
const (
	NV_MAJOR_DEVICE_NUMBER          = 195 // from kernel-open/common/inc/nv.h
	NV_CONTROL_DEVICE_MINOR         = 255 // from kernel-open/common/inc/nv-linux.h
	NVIDIA_UVM_PRIMARY_MINOR_NUMBER = 0   // from kernel-open/nvidia-uvm/uvm_common.h
)

// Handle is NvHandle, from src/common/sdk/nvidia/inc/nvtypes.h.
//
// +marshal
type Handle struct {
	Val uint32
}

// P64 is NvP64, from src/common/sdk/nvidia/inc/nvtypes.h, except that we
// wrap it in a struct so that it can implement marshal.Marshallable.
//
// +marshal
type P64 struct {
	Val uint64
}

// IsNull returns true if the given pointer is NULL.
func (p P64) IsNull() bool {
	return p.Val == 0
}

// IsNotNull returns true if the given pointer is not NULL.
func (p P64) IsNotNull() bool {
	return p.Val != 0
}

const NV_MAX_DEVICES = 32 // src/common/sdk/nvidia/inc/nvlimits.h

// RS_ACCESS_MASK is RS_ACCESS_MASK, from
// src/common/sdk/nvidia/inc/rs_access.h.
//
// +marshal
type RS_ACCESS_MASK struct {
	Limbs [SDK_RS_ACCESS_MAX_LIMBS]uint32 // RsAccessLimb
}

const SDK_RS_ACCESS_MAX_LIMBS = 1
