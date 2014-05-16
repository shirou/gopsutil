package gopsutil

import (
	"encoding/json"
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

type CPUInfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
	Flags      []string `json:"flags"`
}

func CPUCounts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}

func (c CPUTimesStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func (c CPUInfoStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}
