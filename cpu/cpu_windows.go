// +build windows

package cpu

import (
	"strconv"
	"strings"
	"syscall"
	"time"
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
		Idle:   float64(idle),
		User:   float64(user),
		System: float64(system),
	})
	return ret, nil
}

func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat
	lines, err := common.GetWmic("cpu", "Family,L2CacheSize,Manufacturer,Name,NumberOfLogicalProcessors,ProcessorId,Stepping")
	if err != nil {
		return ret, err
	}
	for i, l := range lines {
		t := strings.Split(l, ",")
		if len(t) < 2 {
			continue
		}
		cache, err := strconv.Atoi(t[2])
		if err != nil {
			cache = 0
		}
		cores, err := strconv.Atoi(t[5])
		if err != nil {
			cores = 0
		}
		stepping, err := strconv.Atoi(t[7])
		if err != nil {
			stepping = 0
		}
		cpu := CPUInfoStat{
			CPU:        int32(i),
			Family:     t[1],
			CacheSize:  int32(cache),
			VendorID:   t[3],
			ModelName:  t[4],
			Cores:      int32(cores),
			PhysicalID: t[6],
			Stepping:   int32(stepping),
			Flags:      []string{},
		}
		ret = append(ret, cpu)
	}
	return ret, nil
}

func CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	ret := []float64{}

	lines, err := common.GetWmic("cpu", "loadpercentage")
	if err != nil {
		return ret, err
	}
	for _, l := range lines {
		t := strings.Split(l, ",")

		if len(t) < 2 {
			continue
		}
		p, err := strconv.Atoi(t[1])
		if err != nil {
			p = 0
		}
		ret = append(ret, float64(p)/100.0)
	}
	return ret, nil
}
