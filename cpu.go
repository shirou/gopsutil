package gopsutil

import (
	"runtime"
)

type CPU_TimesStat struct {
	Cpu        string  `json:"cpu"`
	User       float32 `json:"user"`
	System     float32 `json:"system"`
	Idle       float32 `json:"idle"`
	Nice       float32 `json:"nice"`
	Iowait     float32 `json:"iowait"`
	Irq        float32 `json:"irq"`
	Softirq    float32 `json:"softirq"`
	Steal      float32 `json:"steal"`
	Guest      float32 `json:"guest"`
	Guest_nice float32 `json:"guest_nice"`
	Stolen     float32 `json:"stolen"`
}

func Cpu_counts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}
