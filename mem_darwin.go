// +build darwin

package gopsutil

import (
	"os/exec"
	"strconv"
	"strings"
)

func getPageSize() (uint64, error) {
	out, err := exec.Command("pagesize").Output()
	if err != nil {
		return 0, err
	}
	p := mustParseUint64(string(out))
	return p, nil
}

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	p, _ := getPageSize()

	total, _ := doSysctrl("hw.memsize")
	free, _ := doSysctrl("vm.page_free_count")
	/*
		active, _ := doSysctrl("vm.stats.vm.v_active_count")
		inactive, _ := doSysctrl("vm.pageout_inactive_used")
		cache, _ := doSysctrl("vm.stats.vm.v_cache_count")
		buffer, _ := doSysctrl("vfs.bufspace")
		wired, _ := doSysctrl("vm.stats.vm.v_wire_count")
	*/

	ret := &VirtualMemoryStat{
		Total: mustParseUint64(total[0]) * p,
		Free:  mustParseUint64(free[0]) * p,
		/*
			Active:   mustParseUint64(active[0]) * p,
			Inactive: mustParseUint64(inactive[0]) * p,
			Cached:   mustParseUint64(cache[0]) * p,
			Buffers:  mustParseUint64(buffer[0]),
			Wired:    mustParseUint64(wired[0]) * p,
		*/
	}

	// TODO: platform independent (worked freebsd?)
	ret.Available = ret.Free + ret.Buffers + ret.Cached

	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	return ret, nil
}

// SwapMemory returns swapinfo.
func SwapMemory() (*SwapMemoryStat, error) {
	var ret *SwapMemoryStat

	swapUsage, err := doSysctrl("vm.swapusage")
	if err != nil {
		return ret, err
	}

	total := strings.Replace(swapUsage[2], "M", "", 1)
	used := strings.Replace(swapUsage[5], "M", "", 1)
	free := strings.Replace(swapUsage[8], "M", "", 1)

	total_v, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return nil, err
	}
	used_v, err := strconv.ParseFloat(used, 64)
	if err != nil {
		return nil, err
	}
	free_v, err := strconv.ParseFloat(free, 64)
	if err != nil {
		return nil, err
	}

	u := ((total_v - free_v) / total_v) * 100.0

	// vm.swapusage shows "M", multiply 1000
	ret = &SwapMemoryStat{
		Total:       uint64(total_v * 1000),
		Used:        uint64(used_v * 1000),
		Free:        uint64(free_v * 1000),
		UsedPercent: u,
	}

	return ret, nil
}
