package gopsutil

import (
	"runtime"
)

type CPU_TimesStat struct {
	Cpu        string `json:"cpu"`
	User       uint64 `json:"user"`
	System     uint64 `json:"system"`
	Idle       uint64 `json:"idle"`
	Nice       uint64 `json:"nice"`
	Iowait     uint64 `json:"iowait"`
	Irq        uint64 `json:"irq"`
	Softirq    uint64 `json:"softirq"`
	Steal      uint64 `json:"steal"`
	Guest      uint64 `json:"guest"`
	Guest_nice uint64 `json:"guest_nice"`
	Stolen     uint64 `json:"stolen"`
}

func Cpu_counts() (int, error) {
	return runtime.NumCPU(), nil
}
