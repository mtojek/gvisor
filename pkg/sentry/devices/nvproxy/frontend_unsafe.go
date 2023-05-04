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
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/abi/nvgpu"
	"gvisor.dev/gvisor/pkg/errors/linuxerr"
)

func frontendIoctlInvoke[Params any](fi *frontendIoctlState, sentryParams *Params) (uintptr, error) {
	n, _, errno := unix.RawSyscall(unix.SYS_IOCTL, uintptr(fi.fd.hostFD), frontendIoctlCmd(fi.nr, fi.ioctlParamsSize), uintptr(unsafe.Pointer(sentryParams)))
	if errno != 0 {
		return n, errno
	}
	return n, nil
}

func rmControlSimple(fi *frontendIoctlState, ioctlParams *nvgpu.NVOS54Parameters) (uintptr, error) {
	if ioctlParams.ParamsSize == 0 {
		if ioctlParams.Params.IsNotNull() {
			return 0, linuxerr.EINVAL
		}
		return rmControlInvoke[byte](fi, ioctlParams, nil)
	}
	if ioctlParams.Params.IsNull() {
		return 0, linuxerr.EINVAL
	}

	ctrlParams := make([]byte, ioctlParams.ParamsSize)
	if _, err := fi.t.CopyInBytes(addrFromP64(ioctlParams.Params), ctrlParams); err != nil {
		return 0, err
	}
	n, err := rmControlInvoke(fi, ioctlParams, &ctrlParams[0])
	if err != nil {
		return n, err
	}
	if _, err := fi.t.CopyOutBytes(addrFromP64(ioctlParams.Params), ctrlParams); err != nil {
		return n, err
	}
	return n, nil
}

func rmControlInvoke[Params any](fi *frontendIoctlState, ioctlParams *nvgpu.NVOS54Parameters, ctrlParams *Params) (uintptr, error) {
	defer runtime.KeepAlive(ctrlParams) // since we convert to non-pointer-typed P64
	sentryIoctlParams := nvgpu.NVOS54Parameters{
		HClient:    ioctlParams.HClient,
		HObject:    ioctlParams.HObject,
		Cmd:        ioctlParams.Cmd,
		Flags:      ioctlParams.Flags,
		Params:     p64FromPtr(unsafe.Pointer(ctrlParams)),
		ParamsSize: ioctlParams.ParamsSize,
		Status:     ioctlParams.Status,
	}
	n, err := frontendIoctlInvoke(fi, &sentryIoctlParams)
	if err != nil {
		return n, err
	}
	outIoctlParams := nvgpu.NVOS54Parameters{
		HClient:    sentryIoctlParams.HClient,
		HObject:    sentryIoctlParams.HObject,
		Cmd:        sentryIoctlParams.Cmd,
		Flags:      sentryIoctlParams.Flags,
		Params:     ioctlParams.Params,
		ParamsSize: sentryIoctlParams.ParamsSize,
		Status:     sentryIoctlParams.Status,
	}
	if _, err := outIoctlParams.CopyOut(fi.t, fi.ioctlParamsAddr); err != nil {
		return n, err
	}
	return n, nil
}

func ctrlClientSystemGetBuildVersion(fi *frontendIoctlState, ioctlParams *nvgpu.NVOS54Parameters) (uintptr, error) {
	var gbvParams nvgpu.NV0000_CTRL_SYSTEM_GET_BUILD_VERSION_PARAMS
	if unsafe.Sizeof(gbvParams) != uintptr(ioctlParams.ParamsSize) {
		return 0, linuxerr.EINVAL
	}
	if _, err := gbvParams.CopyIn(fi.t, addrFromP64(ioctlParams.Params)); err != nil {
		return 0, err
	}

	if gbvParams.PDriverVersionBuffer.IsNull() || gbvParams.PVersionBuffer.IsNull() || gbvParams.PTitleBuffer.IsNull() {
		// No strings are written if any are null. See
		// src/nvidia/interface/deprecated/rmapi_deprecated_control.c:V2_CONVERTER(_NV0000_CTRL_CMD_SYSTEM_GET_BUILD_VERSION).
		return ctrlClientSystemGetBuildVersionInvoke(fi, ioctlParams, &gbvParams, nil, nil, nil)
	}

	// Need to buffer strings for copy-out.
	if gbvParams.SizeOfStrings == 0 {
		return 0, linuxerr.EINVAL
	}
	driverVersionBuf := make([]byte, gbvParams.SizeOfStrings)
	versionBuf := make([]byte, gbvParams.SizeOfStrings)
	titleBuf := make([]byte, gbvParams.SizeOfStrings)
	n, err := ctrlClientSystemGetBuildVersionInvoke(fi, ioctlParams, &gbvParams, &driverVersionBuf[0], &versionBuf[0], &titleBuf[0])
	if err != nil {
		return n, err
	}
	if _, err := fi.t.CopyOutBytes(addrFromP64(gbvParams.PDriverVersionBuffer), driverVersionBuf); err != nil {
		return n, err
	}
	if _, err := fi.t.CopyOutBytes(addrFromP64(gbvParams.PVersionBuffer), versionBuf); err != nil {
		return n, err
	}
	if _, err := fi.t.CopyOutBytes(addrFromP64(gbvParams.PTitleBuffer), titleBuf); err != nil {
		return n, err
	}
	return n, nil
}

func ctrlClientSystemGetBuildVersionInvoke(fi *frontendIoctlState, ioctlParams *nvgpu.NVOS54Parameters, gbvParams *nvgpu.NV0000_CTRL_SYSTEM_GET_BUILD_VERSION_PARAMS, driverVersionBuf, versionBuf, titleBuf *byte) (uintptr, error) {
	sentryGBVParams := nvgpu.NV0000_CTRL_SYSTEM_GET_BUILD_VERSION_PARAMS{
		SizeOfStrings:            gbvParams.SizeOfStrings,
		PDriverVersionBuffer:     p64FromPtr(unsafe.Pointer(driverVersionBuf)),
		PVersionBuffer:           p64FromPtr(unsafe.Pointer(versionBuf)),
		PTitleBuffer:             p64FromPtr(unsafe.Pointer(titleBuf)),
		ChangelistNumber:         gbvParams.ChangelistNumber,
		OfficialChangelistNumber: gbvParams.OfficialChangelistNumber,
	}
	n, err := rmControlInvoke(fi, ioctlParams, &sentryGBVParams)
	if err != nil {
		return n, err
	}
	outGBVParams := nvgpu.NV0000_CTRL_SYSTEM_GET_BUILD_VERSION_PARAMS{
		SizeOfStrings:            sentryGBVParams.SizeOfStrings,
		PDriverVersionBuffer:     gbvParams.PDriverVersionBuffer,
		PVersionBuffer:           gbvParams.PVersionBuffer,
		PTitleBuffer:             gbvParams.PTitleBuffer,
		ChangelistNumber:         sentryGBVParams.ChangelistNumber,
		OfficialChangelistNumber: sentryGBVParams.OfficialChangelistNumber,
	}
	if _, err := outGBVParams.CopyOut(fi.t, addrFromP64(ioctlParams.Params)); err != nil {
		return n, err
	}
	return n, nil
}

func ctrlSubdevGRGetInfo(fi *frontendIoctlState, ioctlParams *nvgpu.NVOS54Parameters) (uintptr, error) {
	var infoParams nvgpu.NV2080_CTRL_CMD_GR_GET_INFO_PARAMS
	if unsafe.Sizeof(infoParams) != uintptr(ioctlParams.ParamsSize) {
		return 0, linuxerr.EINVAL
	}
	if _, err := infoParams.CopyIn(fi.t, addrFromP64(ioctlParams.Params)); err != nil {
		return 0, err
	}

	if infoParams.GRInfoListSize == 0 {
		// Compare
		// src/nvidia/src/kernel/gpu/gr/kernel_graphics.c:_kgraphicsCtrlCmdGrGetInfoV2().
		return 0, linuxerr.EINVAL
	}
	infoList := make([]byte, uintptr(infoParams.GRInfoListSize)*unsafe.Sizeof(nvgpu.NVXXXX_CTRL_XXX_INFO{}))
	if _, err := fi.t.CopyInBytes(addrFromP64(infoParams.GRInfoList), infoList); err != nil {
		return 0, err
	}

	sentryInfoParams := nvgpu.NV2080_CTRL_CMD_GR_GET_INFO_PARAMS{
		GRInfoListSize: infoParams.GRInfoListSize,
		GRInfoList:     p64FromPtr(unsafe.Pointer(&infoList[0])),
		GRRouteInfo:    infoParams.GRRouteInfo,
	}
	n, err := rmControlInvoke(fi, ioctlParams, &sentryInfoParams)
	if err != nil {
		return n, err
	}

	if _, err := fi.t.CopyOutBytes(addrFromP64(infoParams.GRInfoList), infoList); err != nil {
		return n, err
	}

	outInfoParams := nvgpu.NV2080_CTRL_CMD_GR_GET_INFO_PARAMS{
		GRInfoListSize: sentryInfoParams.GRInfoListSize,
		GRInfoList:     infoParams.GRInfoList,
		GRRouteInfo:    sentryInfoParams.GRRouteInfo,
	}
	if _, err := outInfoParams.CopyOut(fi.t, addrFromP64(ioctlParams.Params)); err != nil {
		return n, err
	}

	return n, nil
}

func rmAllocInvoke[Params any](fi *frontendIoctlState, ioctlParams *nvgpu.NVOS64Parameters, allocParams *Params, isNVOS64 bool) (uintptr, error) {
	defer runtime.KeepAlive(allocParams) // since we convert to non-pointer-typed P64

	if isNVOS64 {
		sentryIoctlParams := nvgpu.NVOS64Parameters{
			HRoot:         ioctlParams.HRoot,
			HObjectParent: ioctlParams.HObjectParent,
			HObjectNew:    ioctlParams.HObjectNew,
			HClass:        ioctlParams.HClass,
			PAllocParms:   p64FromPtr(unsafe.Pointer(allocParams)),
			Flags:         ioctlParams.Flags,
			Status:        ioctlParams.Status,
		}
		var rightsRequested nvgpu.RS_ACCESS_MASK
		if ioctlParams.PRightsRequested.IsNotNull() {
			if _, err := rightsRequested.CopyIn(fi.t, addrFromP64(ioctlParams.PRightsRequested)); err != nil {
				return 0, err
			}
			sentryIoctlParams.PRightsRequested = p64FromPtr(unsafe.Pointer(&rightsRequested))
		}
		n, err := frontendIoctlInvoke(fi, &sentryIoctlParams)
		if err != nil {
			return n, err
		}
		if ioctlParams.PRightsRequested.IsNotNull() {
			if _, err := rightsRequested.CopyOut(fi.t, addrFromP64(ioctlParams.PRightsRequested)); err != nil {
				return n, err
			}
		}
		outIoctlParams := nvgpu.NVOS64Parameters{
			HRoot:            sentryIoctlParams.HRoot,
			HObjectParent:    sentryIoctlParams.HObjectParent,
			HObjectNew:       sentryIoctlParams.HObjectNew,
			HClass:           sentryIoctlParams.HClass,
			PAllocParms:      ioctlParams.PAllocParms,
			PRightsRequested: ioctlParams.PRightsRequested,
			Flags:            sentryIoctlParams.Flags,
			Status:           sentryIoctlParams.Status,
		}
		if _, err := outIoctlParams.CopyOut(fi.t, fi.ioctlParamsAddr); err != nil {
			return n, err
		}
		return n, nil
	}

	sentryIoctlParams := nvgpu.NVOS21Parameters{
		HRoot:         ioctlParams.HRoot,
		HObjectParent: ioctlParams.HObjectParent,
		HObjectNew:    ioctlParams.HObjectNew,
		HClass:        ioctlParams.HClass,
		PAllocParms:   p64FromPtr(unsafe.Pointer(allocParams)),
		Status:        ioctlParams.Status,
	}
	n, err := frontendIoctlInvoke(fi, &sentryIoctlParams)
	if err != nil {
		return n, err
	}
	outIoctlParams := nvgpu.NVOS21Parameters{
		HRoot:         sentryIoctlParams.HRoot,
		HObjectParent: sentryIoctlParams.HObjectParent,
		HObjectNew:    sentryIoctlParams.HObjectNew,
		HClass:        sentryIoctlParams.HClass,
		PAllocParms:   ioctlParams.PAllocParms,
		Status:        sentryIoctlParams.Status,
	}
	if _, err := outIoctlParams.CopyOut(fi.t, fi.ioctlParamsAddr); err != nil {
		return n, err
	}
	return n, nil
}
