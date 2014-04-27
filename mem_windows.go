// +build windows

package gopsutil

import (
	"syscall"
	"unsafe"
)

var (
	procGlobalMemoryStatusEx = modkernel32.NewProc("GlobalMemoryStatusEx")
)

type MEMORYSTATUSEX struct {
	cbSize                  uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64 // in bytes
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func Virtual_memory() (Virtual_memoryStat, error) {
	ret := Virtual_memoryStat{}

	var memInfo MEMORYSTATUSEX
	memInfo.cbSize = uint32(unsafe.Sizeof(memInfo))
	mem, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memInfo)))
	if mem == 0 {
		return ret, syscall.GetLastError()
	}

	ret.Total = memInfo.ullTotalPhys
	ret.Available = memInfo.ullAvailPhys
	ret.UsedPercent = float64(memInfo.dwMemoryLoad)
	ret.Used = ret.Total - ret.Available
	return ret, nil
}

func Swap_memory() (Swap_memoryStat, error) {
	ret := Swap_memoryStat{}

	return ret, nil
}
