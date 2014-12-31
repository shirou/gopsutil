// +build windows

package cpu

import (
	"syscall"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

// TODO: Get percpu
func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	var ret []CPUTimesStat

	var lpIdleTime common.FILETIME
	var lpKernelTime common.FILETIME
	var lpUserTime common.FILETIME
	r, _, _ := common.ProcGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)))
	if r == 0 {
		return ret, syscall.GetLastError()
	}

	LOT := float64(0.0000001)
	HIT := (LOT * 4294967296.0)
	idle := ((HIT * float64(lpIdleTime.DwHighDateTime)) + (LOT * float64(lpIdleTime.DwLowDateTime)))
	user := ((HIT * float64(lpUserTime.DwHighDateTime)) + (LOT * float64(lpUserTime.DwLowDateTime)))
	kernel := ((HIT * float64(lpKernelTime.DwHighDateTime)) + (LOT * float64(lpKernelTime.DwLowDateTime)))
	system := (kernel - idle)

	ret = append(ret, CPUTimesStat{
		Idle:   float32(idle),
		User:   float32(user),
		System: float32(system),
	})
	return ret, nil
}

func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat
	return ret, nil
}
