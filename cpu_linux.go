// +build linux

package gopsutil

import (
	"strconv"
	"strings"
)

func Cpu_times() ([]CPU_TimesStat, error) {
	ret := make([]CPU_TimesStat, 0)

	filename := "/proc/stat"
	lines, _ := ReadLines(filename)
	for _, line := range lines {
		fields := strings.Fields(line)

		if strings.HasPrefix(fields[0], "cpu") == false {
			continue
		}

		cpu := fields[0]
		if cpu == "cpu" {
			cpu = "cpu-total"
		}
		user, _ := strconv.ParseUint(fields[1], 10, 64)
		nice, _ := strconv.ParseUint(fields[2], 10, 64)
		system, _ := strconv.ParseUint(fields[3], 10, 64)
		idle, _ := strconv.ParseUint(fields[4], 10, 64)
		iowait, _ := strconv.ParseUint(fields[5], 10, 64)
		irq, _ := strconv.ParseUint(fields[6], 10, 64)
		softirq, _ := strconv.ParseUint(fields[7], 10, 64)
		stolen, _ := strconv.ParseUint(fields[8], 10, 64)
		ct := CPU_TimesStat{
			Cpu:     cpu,
			User:    user,
			Nice:    nice,
			System:  system,
			Idle:    idle,
			Iowait:  iowait,
			Irq:     irq,
			Softirq: softirq,
			Stolen:  stolen,
		}
		if len(fields) > 9 { // Linux >= 2.6.11
			steal, _ := strconv.ParseUint(fields[9], 10, 64)
			ct.Steal = steal
		}
		if len(fields) > 10 { // Linux >= 2.6.24
			guest, _ := strconv.ParseUint(fields[10], 10, 64)
			ct.Guest = guest
		}
		if len(fields) > 11 { // Linux >= 3.2.0
			guest_nice, _ := strconv.ParseUint(fields[11], 10, 64)
			ct.Guest_nice = guest_nice
		}

		ret = append(ret, ct)
	}

	return ret, nil
}
