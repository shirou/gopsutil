// +build darwin

package gopsutil

import (
	"os/exec"
	"strconv"
	"strings"
)

// sys/resource.h
const (
	CP_USER   = 0
	CP_NICE   = 1
	CP_SYS    = 2
	CP_INTR   = 3
	CP_IDLE   = 4
	CPUSTATES = 5
)

// time.h
const (
	CLOCKS_PER_SEC = 128
)

// TODO: get per cpus
func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	var ret []CPUTimesStat

	cpuTime, err := doSysctrl("kern.cp_time")
	if err != nil {
		return ret, err
	}

	user, _ := strconv.ParseFloat(cpuTime[CP_USER], 32)
	nice, _ := strconv.ParseFloat(cpuTime[CP_NICE], 32)
	sys, _ := strconv.ParseFloat(cpuTime[CP_SYS], 32)
	idle, _ := strconv.ParseFloat(cpuTime[CP_IDLE], 32)
	intr, _ := strconv.ParseFloat(cpuTime[CP_INTR], 32)

	c := CPUTimesStat{
		User:   float32(user / CLOCKS_PER_SEC),
		Nice:   float32(nice / CLOCKS_PER_SEC),
		System: float32(sys / CLOCKS_PER_SEC),
		Idle:   float32(idle / CLOCKS_PER_SEC),
		Irq:    float32(intr / CLOCKS_PER_SEC), // FIXME: correct?
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
			c.Stepping = mustParseInt32(values[1])
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
			c.Cores = mustParseInt32(values[1])
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			c.CacheSize = mustParseInt32(values[1])
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}

		// TODO:
		// c.Mhz = mustParseFloat64(values[1])
	}

	return append(ret, c), nil
}
