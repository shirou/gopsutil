// +build freebsd

package gopsutil

import (
	"regexp"
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
	filename := "/var/run/dmesg.boot"
	lines, _ := readLines(filename)

	var ret []CPUInfoStat

	c := CPUInfoStat{}
	for _, line := range lines {
		if matches := regexp.MustCompile(`CPU:\s+(.+) \(([\d.]+).+\)`).FindStringSubmatch(line); matches != nil {
			c.ModelName = matches[1]
			t, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return ret, nil
			}
			c.Mhz = t
		} else if matches := regexp.MustCompile(`Origin = "(.+)"  Id = (.+)  Family = (.+)  Model = (.+)  Stepping = (.+)`).FindStringSubmatch(line); matches != nil {
			c.VendorID = matches[1]
			c.Family = matches[3]
			c.Model = matches[4]
			t, err := strconv.ParseInt(matches[5], 10, 32)
			if err != nil {
				return ret, nil
			}
			c.Stepping = int32(t)
		} else if matches := regexp.MustCompile(`Features=.+<(.+)>`).FindStringSubmatch(line); matches != nil {
			for _, v := range strings.Split(matches[1], ",") {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if matches := regexp.MustCompile(`Features2=[a-f\dx]+<(.+)>`).FindStringSubmatch(line); matches != nil {
			for _, v := range strings.Split(matches[1], ",") {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if matches := regexp.MustCompile(`Logical CPUs per core: (\d+)`).FindStringSubmatch(line); matches != nil {
			// FIXME: no this line?
			t, err := strconv.ParseInt(matches[1], 10, 32)
			if err != nil {
				return ret, nil
			}
			c.Cores = int32(t)
		}

	}

	return append(ret, c), nil
}
