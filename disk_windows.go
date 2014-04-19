// +build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	procGetDiskFreeSpaceExW = modkernel32.NewProc("GetDiskFreeSpaceExW")
	//GetLogicalDriveStrings, _ = syscall.GetProcAddress(modkernel32, "GetLogicalDriveStringsW")
)

func (d Disk) Disk_usage(path string) (Disk_usage, error) {
	ret := Disk_usage{}

	ret.Path = path
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	diskret, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	if diskret == 0 {
		return ret, syscall.GetLastError()
	}
	ret.Total = uint64(lpTotalNumberOfBytes)
	//	ret.Free = uint64(lpFreeBytesAvailable) // python psutil does not use this
	ret.Free = uint64(lpTotalNumberOfFreeBytes)
	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Used) / float64(ret.Total) * 100.0

	return ret, nil
}
