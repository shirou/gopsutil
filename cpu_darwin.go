// +build darwin

package gopsutil

import (
	"os/exec"
	"strconv"
	"strings"
)

// sys/resource.h
const (
	CPUser    = 0
	CPNice    = 1
	CPSys     = 2
	CPIntr    = 3
	CPIdle    = 4
	CPUStates = 5
)

// time.h
const (
	ClocksPerSec = 128
)

// TODO: get per cpus
func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	var ret []CPUTimesStat

	cpuTime, err := doSysctrl("kern.cp_time")
	if err != nil {
		return ret, err
	}

	user, err := strconv.ParseFloat(cpuTime[CPUser], 32)
	if err != nil {
		return ret, err
	}
	nice, err := strconv.ParseFloat(cpuTime[CPNice], 32)
	if err != nil {
		return ret, err
	}
	sys, err := strconv.ParseFloat(cpuTime[CPSys], 32)
	if err != nil {
		return ret, err
	}
	idle, err := strconv.ParseFloat(cpuTime[CPIdle], 32)
	if err != nil {
		return ret, err
	}
	intr, err := strconv.ParseFloat(cpuTime[CPIntr], 32)
	if err != nil {
		return ret, err
	}

	c := CPUTimesStat{
		User:   float32(user / ClocksPerSec),
		Nice:   float32(nice / ClocksPerSec),
		System: float32(sys / ClocksPerSec),
		Idle:   float32(idle / ClocksPerSec),
		Irq:    float32(intr / ClocksPerSec),
	}

	ret = append(ret, c)

	return ret, nil
}

// Returns only one CPUInfoStat on FreeBSD
func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat

	out, err := exec.Command("/usr/sbin/sysctl", "machdep.cpu").Output()
	if err != nil {
		return ret, err
	}

	c := CPUInfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)

		if strings.HasPrefix(line, "machdep.cpu.brand_string") {
			c.ModelName = strings.Join(values[1:], " ")
		} else if strings.HasPrefix(line, "machdep.cpu.family") {
			c.Family = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.model") {
			c.Model = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.stepping") {
			t, err := strconv.ParseInt(values[1], 10, 32)
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)

		} else if strings.HasPrefix(line, "machdep.cpu.features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.leaf7_features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.extfeatures") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.core_count") {
			t, err := strconv.ParseInt(values[1], 10, 32)
			if err != nil {
				return ret, err
			}
			c.Cores = t
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			t, err := strconv.ParseInt(values[1], 10, 32)
			if err != nil {
				return ret, err
			}

			c.CacheSize = t
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}

		// TODO:
		// c.Mhz = mustParseFloat64(values[1])
	}

	return append(ret, c), nil
}
