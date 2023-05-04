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

// Package nvproxy implements proxying for the Nvidia GPU Linux kernel driver:
// https://github.com/NVIDIA/open-gpu-kernel-modules
package nvproxy

import (
	"fmt"

	"gvisor.dev/gvisor/pkg/abi/nvgpu"
	"gvisor.dev/gvisor/pkg/context"
	"gvisor.dev/gvisor/pkg/hostarch"
	"gvisor.dev/gvisor/pkg/marshal"
	"gvisor.dev/gvisor/pkg/sentry/fsimpl/devtmpfs"
	"gvisor.dev/gvisor/pkg/sentry/vfs"
)

// Register registers all devices implemented by this package in vfsObj.
func Register(vfsObj *vfs.VirtualFilesystem) (uvmDevMajor uint32, err error) {
	udm, err := vfsObj.GetDynamicCharDevMajor()
	if err != nil {
		return 0, err
	}

	nvp := &nvproxy{}
	for minor := uint32(0); minor <= nvgpu.NV_CONTROL_DEVICE_MINOR; minor++ {
		if err := vfsObj.RegisterDevice(vfs.CharDevice, nvgpu.NV_MAJOR_DEVICE_NUMBER, minor, &frontendDevice{
			nvp:   nvp,
			minor: minor,
		}, &vfs.RegisterDeviceOptions{
			GroupName: "nvidia-frontend",
		}); err != nil {
			return 0, err
		}
	}
	if err := vfsObj.RegisterDevice(vfs.CharDevice, udm, nvgpu.NVIDIA_UVM_PRIMARY_MINOR_NUMBER, &uvmDevice{
		nvp: nvp,
	}, &vfs.RegisterDeviceOptions{
		GroupName: "nvidia-uvm",
	}); err != nil {
		return 0, err
	}
	return udm, nil
}

// CreateDriverDevtmpfsFiles creates device special files in dev that should
// always exist when this package is enabled. It does not create per-device
// files in dev; see CreateIndexDevtmpfsFile.
func CreateDriverDevtmpfsFiles(ctx context.Context, dev *devtmpfs.Accessor, uvmDevMajor uint32) error {
	if err := dev.CreateDeviceFile(ctx, "nvidiactl", vfs.CharDevice, nvgpu.NV_MAJOR_DEVICE_NUMBER, nvgpu.NV_CONTROL_DEVICE_MINOR, 0666); err != nil {
		return err
	}
	if err := dev.CreateDeviceFile(ctx, "nvidia-uvm", vfs.CharDevice, uvmDevMajor, nvgpu.NVIDIA_UVM_PRIMARY_MINOR_NUMBER, 0666); err != nil {
		return err
	}
	return nil
}

// CreateIndexDevtmpfsFile creates the device special file in dev for the
// device with the given index.
func CreateIndexDevtmpfsFile(ctx context.Context, dev *devtmpfs.Accessor, index uint32) error {
	return dev.CreateDeviceFile(ctx, fmt.Sprintf("nvidia%d", index), vfs.CharDevice, nvgpu.NV_MAJOR_DEVICE_NUMBER, index, 0666)
}

// +stateify savable
type nvproxy struct {
}

type marshalPtr[T any] interface {
	*T
	marshal.Marshallable
}

func addrFromP64(p nvgpu.P64) hostarch.Addr {
	return hostarch.Addr(p.Val)
}
