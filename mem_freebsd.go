// +build freebsd

package gopsutil

import (
	"os/exec"
	"strings"
)

func VirtualMemory() (*VirtualMemoryStat, error) {

	pageSize, _ := doSysctrl("vm.stats.vm.v_page_size")
	p := mustParseUint64(pageSize[0])

	pageCount, _ := doSysctrl("vm.stats.vm.v_page_count")
	free, _ := doSysctrl("vm.stats.vm.v_free_count")
	active, _ := doSysctrl("vm.stats.vm.v_active_count")
	inactive, _ := doSysctrl("vm.stats.vm.v_inactive_count")
	cache, _ := doSysctrl("vm.stats.vm.v_cache_count")
	buffer, _ := doSysctrl("vfs.bufspace")
	wired, _ := doSysctrl("vm.stats.vm.v_wire_count")

	ret := &VirtualMemoryStat{
		Total:    mustParseUint64(pageCount[0]) * p,
		Free:     mustParseUint64(free[0]) * p,
		Active:   mustParseUint64(active[0]) * p,
		Inactive: mustParseUint64(inactive[0]) * p,
		Cached:   mustParseUint64(cache[0]) * p,
		Buffers:  mustParseUint64(buffer[0]),
		Wired:    mustParseUint64(wired[0]) * p,
	}

	// TODO: platform independent (worked freebsd?)
	ret.Available = ret.Free + ret.Buffers + ret.Cached

	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	return ret, nil
}

// Return swapinfo
// FreeBSD can have multiple swap devices. but use only first device
func SwapMemory() (*SwapMemoryStat, error) {
	out, err := exec.Command("swapinfo").Output()
	if err != nil {
		return nil, err
	}
	var ret *SwapMemoryStat
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		// skip title line
		if len(values) == 0 || values[0] == "Device" {
			continue
		}

		u := strings.Replace(values[4], "%", "", 1)

		ret = &SwapMemoryStat{
			Total:       mustParseUint64(values[1]),
			Used:        mustParseUint64(values[2]),
			Free:        mustParseUint64(values[3]),
			UsedPercent: mustParseFloat64(u),
		}
	}

	return ret, nil
}
