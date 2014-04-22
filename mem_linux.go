// +build freebsd linux

package gopsutil

import (
	"syscall"
)

func (m Mem) Virtual_memory() (Virtual_memory, error) {
	ret := Virtual_memory{}
	sysinfo := &syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(sysinfo); err != nil {
		return ret, err
	}
	ret.Total = uint64(sysinfo.Totalram)
	ret.Free = uint64(sysinfo.Freeram)
	ret.Shared = uint64(sysinfo.Sharedram)
	ret.Buffers = uint64(sysinfo.Bufferram)

	ret.Used = ret.Total - ret.Free

	// TODO: platform independent
	ret.Available = ret.Free + ret.Buffers + ret.Cached

	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	/*
		kern := buffers + cached
		ret.ActualFree = ret.Free + kern
		ret.ActualUsed = ret.Used - kern
	*/

	return ret, nil
}

func (m Mem) Swap_memory() (Swap_memory, error) {
	ret := Swap_memory{}
	sysinfo := &syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(sysinfo); err != nil {
		return ret, err
	}
	ret.Total = sysinfo.Totalswap
	ret.Free = sysinfo.Freeswap
	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Free) / float64(ret.Total) * 100.0

	return ret, nil
}
