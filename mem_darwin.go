// +build darwin

package gopsutil

import (
	"os/exec"
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
	swapUsage, _ := doSysctrl("vm.swapusage")

	var ret *SwapMemoryStat

	total := strings.Replace(swapUsage[3], "M", "", 1)
	used := strings.Replace(swapUsage[6], "M", "", 1)
	free := strings.Replace(swapUsage[9], "M", "", 1)

	u := "0"

	ret = &SwapMemoryStat{
		Total:       mustParseUint64(total),
		Used:        mustParseUint64(used),
		Free:        mustParseUint64(free),
		UsedPercent: mustParseFloat64(u),
	}

	return ret, nil
}
