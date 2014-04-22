// +build windows

package gopsutil

import (
	"syscall"
	"unsafe"
)

var (
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
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

func (m Mem) Virtual_memory() (Virtual_memory, error) {
	ret := Virtual_memory{}

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

func (m Mem) Swap_memory() (Swap_memory, error) {
	ret := Swap_memory{}

	return ret, nil
}
