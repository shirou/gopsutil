// +build linux

package gopsutil

import (
	"strconv"
	"strings"
	"syscall"
)

func VirtualMemory() (*VirtualMemoryStat, error) {
	filename := "/proc/meminfo"
	lines, _ := readLines(filename)

	ret := &VirtualMemoryStat{}
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		value = strings.Replace(value, " kB", "", -1)

		switch key {
		case "MemTotal":
			ret.Total = mustParseUint64(value) * 1000
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Total = uint64(t) * 1000
		case "MemFree":
			ret.Free = mustParseUint64(value) * 1000
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Free = uint64(t) * 1000
		case "Buffers":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Buffers = uint64(t) * 1000
		case "Cached":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Cached = uint64(t) * 1000
		case "Active":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Active = uint64(t) * 1000
		case "Inactive":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			ret.Inactive = uint64(t) * 1000
		}
	}
	ret.Available = ret.Free + ret.Buffers + ret.Cached
	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	return ret, nil
}

func SwapMemory() (*SwapMemoryStat, error) {
	sysinfo := &syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(sysinfo); err != nil {
		return nil, err
	}
	ret := &SwapMemoryStat{
		Total: uint64(sysinfo.Totalswap),
		Free:  uint64(sysinfo.Freeswap),
	}
	ret.Used = ret.Total - ret.Free
	//check Infinity
	if ret.Total != 0 {
		ret.UsedPercent = float64(ret.Total-ret.Free) / float64(ret.Total) * 100.0
	} else {
		ret.UsedPercent = 0
	}

	return ret, nil
}
