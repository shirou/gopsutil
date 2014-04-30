// +build freebsd

package gopsutil

import (
	"strconv"
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
