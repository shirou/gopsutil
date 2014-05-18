// +build freebsd

package gopsutil

import (
	"regexp"
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
	filename := "/var/run/dmesg.boot"
	lines, _ := readLines(filename)

	var ret []CPUInfoStat

	c := CPUInfoStat{}
	for _, line := range lines {
		if matches := regexp.MustCompile(`CPU:\s+(.+) \(([\d.]+).+\)`).FindStringSubmatch(line); matches != nil {
			c.ModelName = matches[1]
			c.Mhz = mustParseFloat64(matches[2])
		} else if matches := regexp.MustCompile(`Origin = "(.+)"  Id = (.+)  Family = (.+)  Model = (.+)  Stepping = (.+)`).FindStringSubmatch(line); matches != nil {
			c.VendorID = matches[1]
			c.Family = matches[3]
			c.Model = matches[4]
			c.Stepping = mustParseInt32(matches[5])
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
			c.Cores = mustParseInt32(matches[1])
		}

	}

	return append(ret, c), nil
}
