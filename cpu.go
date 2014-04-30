package gopsutil

import (
	"runtime"
)

type CPUTimesStat struct {
	CPU       string  `json:"cpu"`
	User      float32 `json:"user"`
	System    float32 `json:"system"`
	Idle      float32 `json:"idle"`
	Nice      float32 `json:"nice"`
	Iowait    float32 `json:"iowait"`
	Irq       float32 `json:"irq"`
	Softirq   float32 `json:"softirq"`
	Steal     float32 `json:"steal"`
	Guest     float32 `json:"guest"`
	GuestNice float32 `json:"guest_nice"`
	Stolen    float32 `json:"stolen"`
}

func CPUCounts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}
