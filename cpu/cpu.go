package cpu

import (
	"encoding/json"
	"runtime"
	"time"
)

type CPUTimesStat struct {
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guest_nice"`
	Stolen    float64 `json:"stolen"`
}

type CPUInfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendor_id"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physical_id"`
	CoreID     string   `json:"core_id"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"model_name"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cache_size"`
	Flags      []string `json:"flags"`
}

func CPUCounts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}

var lastCPUTimes []CPUTimesStat
var lastPerCPUTimes []CPUTimesStat

func CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	getAllBusy := func(t CPUTimesStat) (float64, float64) {
		busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
			t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
		return busy + t.Idle, busy
	}

	calculate := func(t1, t2 CPUTimesStat) float64 {
		t1All, t1Busy := getAllBusy(t1)
		t2All, t2Busy := getAllBusy(t2)

		if t2Busy <= t1Busy {
			return 0
		}
		if t2All <= t1All {
			return 1
		}
		return (t2Busy - t1Busy) / (t2All - t1All) * 100
	}

	cpuTimes, err := CPUTimes(percpu)
	if err != nil {
		return nil, err
	}

	if interval > 0 {
		if !percpu {
			lastCPUTimes = cpuTimes
		} else {
			lastPerCPUTimes = cpuTimes
		}
		time.Sleep(interval)
		cpuTimes, err = CPUTimes(percpu)
		if err != nil {
			return nil, err
		}
	}

	ret := make([]float64, len(cpuTimes))
	if !percpu {
		ret[0] = calculate(lastCPUTimes[0], cpuTimes[0])
		lastCPUTimes = cpuTimes
	} else {
		for i, t := range cpuTimes {
			ret[i] = calculate(lastPerCPUTimes[i], t)
		}
		lastPerCPUTimes = cpuTimes
	}
	return ret, nil
}

func (c CPUTimesStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func (c CPUInfoStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func init() {
	lastCPUTimes, _ = CPUTimes(false)
	lastPerCPUTimes, _ = CPUTimes(true)
}
