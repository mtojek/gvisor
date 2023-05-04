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

// Class handles, from src/nvidia/generated/g_allclasses.h.
const (
	NV01_ROOT          = 0x00000000
	NV01_ROOT_NON_PRIV = 0x00000001
	NV01_ROOT_CLIENT   = 0x00000041
	NV01_DEVICE_0      = 0x00000080
	NV20_SUBDEVICE_0   = 0x00002080
)

// NV0080_ALLOC_PARAMETERS is the alloc params type for NV01_DEVICE_0, from
// src/common/sdk/nvidia/inc/class/cl0080.h.
//
// +marshal
type NV0080_ALLOC_PARAMETERS struct {
	DeviceID        uint32
	HClientShare    Handle
	HTargetClient   Handle
	HTargetDevice   Handle
	Flags           uint32
	Pad             [4]byte
	VASpaceSize     uint64
	VAStartInternal uint64
	VALimitInternal uint64
	VAMode          uint32
}

// NV2080_ALLOC_PARAMETERS is the alloc params type for NV20_SUBDEVICE_0, from
// src/common/sdk/nvidia/inc/class/cl2080.h.
//
// +marshal
type NV2080_ALLOC_PARAMETERS struct {
	SubDeviceID uint32
}
