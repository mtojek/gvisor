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

// From src/common/sdk/nvidia/inc/ctrl/ctrlxxxx.h:

// +marshal
type NVXXXX_CTRL_XXX_INFO struct {
	Index uint32
	Data  uint32
}

// From src/common/sdk/nvidia/inc/ctrl/ctrl0000/ctrl0000client.h:
const (
	NV0000_CTRL_CMD_CLIENT_SET_INHERITED_SHARE_POLICY = 0xd04
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl0000/ctrl0000gpu.h:
const (
	NV0000_CTRL_CMD_GPU_GET_ATTACHED_IDS  = 0x201
	NV0000_CTRL_CMD_GPU_GET_ID_INFO       = 0x202
	NV0000_CTRL_CMD_GPU_GET_ID_INFO_V2    = 0x205
	NV0000_CTRL_CMD_GPU_GET_PROBED_IDS    = 0x214
	NV0000_CTRL_CMD_GPU_ATTACH_IDS        = 0x215
	NV0000_CTRL_CMD_GPU_DETACH_IDS        = 0x216
	NV0000_CTRL_CMD_GPU_GET_PCI_INFO      = 0x21b
	NV0000_CTRL_CMD_GPU_QUERY_DRAIN_STATE = 0x279
	NV0000_CTRL_CMD_GPU_GET_MEMOP_ENABLE  = 0x27b
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl0000/ctrl0000syncgpuboost.h:
const (
	NV0000_CTRL_CMD_SYNC_GPU_BOOST_GROUP_INFO = 0xa04
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl0000/ctrl0000system.h:
const (
	NV0000_CTRL_CMD_SYSTEM_GET_BUILD_VERSION = 0x101
)

// +marshal
type NV0000_CTRL_SYSTEM_GET_BUILD_VERSION_PARAMS struct {
	SizeOfStrings            uint32
	Pad                      [4]byte
	PDriverVersionBuffer     P64
	PVersionBuffer           P64
	PTitleBuffer             P64
	ChangelistNumber         uint32
	OfficialChangelistNumber uint32
}

// From src/common/sdk/nvidia/inc/ctrl/ctrl0080/ctrl0080fb.h:
const (
	NV0080_CTRL_CMD_FB_GET_CAPS_V2 = 0x801307
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl0080/ctrl0080gpu.h:
const (
	NV0080_CTRL_CMD_GPU_GET_NUM_SUBDEVICES         = 0x800280
	NV0080_CTRL_CMD_GPU_QUERY_SW_STATE_PERSISTENCE = 0x800288
	NV0080_CTRL_CMD_GPU_GET_VIRTUALIZATION_MODE    = 0x800289
	NV0080_CTRL_CMD_GPU_GET_CLASSLIST_V2           = 0x800292
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl0080/ctrl0080gr.h:

// +marshal
type NV0080_CTRL_GR_ROUTE_INFO struct {
	Flags uint32
	Pad   [4]byte
	Route uint64
}

// From src/common/sdk/nvidia/inc/ctrl/ctrl0080/ctrl0080host.h:
const (
	NV0080_CTRL_CMD_HOST_GET_CAPS_V2 = 0x801402
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080bus.h:
const (
	NV2080_CTRL_CMD_BUS_GET_PCI_INFO                   = 0x20801801
	NV2080_CTRL_CMD_BUS_GET_PCI_BAR_INFO               = 0x20801803
	NV2080_CTRL_CMD_BUS_GET_INFO_V2                    = 0x20801823
	NV2080_CTRL_CMD_BUS_GET_PCIE_SUPPORTED_GPU_ATOMICS = 0x2080182a
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080ce.h:
const (
	NV2080_CTRL_CMD_CE_GET_ALL_CAPS = 0x20802a0a
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080fb.h:
const (
	NV2080_CTRL_CMD_FB_GET_INFO_V2 = 0x20801303
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080gpu.h:
const (
	NV2080_CTRL_CMD_GPU_GET_INFO_V2               = 0x20800102
	NV2080_CTRL_CMD_GPU_GET_NAME_STRING           = 0x20800110
	NV2080_CTRL_CMD_GPU_GET_SIMULATION_INFO       = 0x20800119
	NV2080_CTRL_CMD_GPU_QUERY_ECC_STATUS          = 0x2080012f
	NV2080_CTRL_CMD_GPU_QUERY_COMPUTE_MODE_RULES  = 0x20800131
	NV2080_CTRL_CMD_GPU_GET_GID_INFO              = 0x2080014a
	NV2080_CTRL_CMD_GPU_GET_ENGINES_V2            = 0x20800170
	NV2080_CTRL_CMD_GPU_GET_ACTIVE_PARTITION_IDS  = 0x2080018b
	NV2080_CTRL_CMD_GPU_GET_COMPUTE_POLICY_CONFIG = 0x20800195
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080gr.h:
const (
	NV2080_CTRL_CMD_GR_GET_INFO            = 0x20801201
	NV2080_CTRL_CMD_GR_GET_GLOBAL_SM_ORDER = 0x2080121b
	NV2080_CTRL_CMD_GR_GET_CAPS_V2         = 0x20801227
	NV2080_CTRL_CMD_GR_GET_GPC_MASK        = 0x2080122a
	NV2080_CTRL_CMD_GR_GET_TPC_MASK        = 0x2080122b
)

// +marshal
type NV2080_CTRL_CMD_GR_GET_INFO_PARAMS struct {
	GRInfoListSize uint32 // in elements
	Pad            [4]byte
	GRInfoList     P64
	GRRouteInfo    NV0080_CTRL_GR_ROUTE_INFO
}

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080mc.h:
const (
	NV2080_CTRL_CMD_MC_GET_ARCH_INFO = 0x20801701
)

// From src/common/sdk/nvidia/inc/ctrl/ctrl2080/ctrl2080tmr.h:
const (
	NV2080_CTRL_CMD_TIMER_GET_GPU_CPU_TIME_CORRELATION_INFO = 0x20800406
)
