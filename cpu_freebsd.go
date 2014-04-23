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
func Cpu_times() ([]CPU_TimesStat, error) {
	ret := make([]CPU_TimesStat, 0)

	cpu_time, err := do_sysctrl("kern.cp_time")
	if err != nil {
		return ret, err
	}

	user, _ := strconv.ParseInt(cpu_time[CP_USER], 10, 64)
	nice, _ := strconv.ParseInt(cpu_time[CP_NICE], 10, 64)
	sys, _ := strconv.ParseInt(cpu_time[CP_SYS], 10, 64)
	idle, _ := strconv.ParseInt(cpu_time[CP_IDLE], 10, 64)
	intr, _ := strconv.ParseInt(cpu_time[CP_INTR], 10, 64)

	c := CPU_TimesStat{
		User:   uint64(user / CLOCKS_PER_SEC),
		Nice:   uint64(nice / CLOCKS_PER_SEC),
		System: uint64(sys / CLOCKS_PER_SEC),
		Idle:   uint64(idle / CLOCKS_PER_SEC),
		Irq:    uint64(intr / CLOCKS_PER_SEC), // FIXME: correct?
	}

	ret = append(ret, c)

	return ret, nil
}
