// +build windows

package gopsutil

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32        = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes = modkernel32.NewProc("GetSystemTimes")
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

func Cpu_times() ([]CPU_TimesStat, error) {
	ret := make([]CPU_TimesStat, 0)

	var lpIdleTime FILETIME
	var lpKernelTime FILETIME
	var lpUserTime FILETIME
	r, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)))
	if r == 0 {
		return ret, syscall.GetLastError()
	}

	LO_T := float64(0.0000001)
	HI_T := (LO_T * 4294967296.0)
	idle := ((HI_T * float64(lpIdleTime.DwHighDateTime)) + (LO_T * float64(lpIdleTime.DwLowDateTime)))
	user := ((HI_T * float64(lpUserTime.DwHighDateTime)) + (LO_T * float64(lpUserTime.DwLowDateTime)))
	kernel := ((HI_T * float64(lpKernelTime.DwHighDateTime)) + (LO_T * float64(lpKernelTime.DwLowDateTime)))
	system := (kernel - idle)

	ret = append(ret, CPU_TimesStat{
		Idle:   uint64(idle),
		User:   uint64(user),
		System: uint64(system),
	})
	return ret, nil
}
