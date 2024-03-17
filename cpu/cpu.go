package cpu

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/internal/common"
)

// TimesStat contains the amounts of time the CPU has spent performing different
// kinds of work. Time units are in seconds. It is based on linux /proc/stat file.
type TimesStat struct {
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
	GuestNice float64 `json:"guestNice"`
}

type InfoStat struct {
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
	Microcode  string   `json:"microcode"`
}

type lastPercent struct {
	sync.Mutex
	lastCPUTimes    []TimesStat
	lastPerCPUTimes []TimesStat
}

var (
	lastCPUPercent lastPercent
	invoke         common.Invoker = common.Invoke{}
)

func init() {
	lastCPUPercent.Lock()
	lastCPUPercent.lastCPUTimes, _ = Times(false)
	lastCPUPercent.lastPerCPUTimes, _ = Times(true)
	lastCPUPercent.Unlock()
}

// Counts returns the number of physical or logical cores in the system
func Counts(logical bool) (int, error) {
	return CountsWithContext(context.Background(), logical)
}

func (c TimesStat) String() string {
	v := []string{
		`"cpu":"` + c.CPU + `"`,
		`"user":` + strconv.FormatFloat(c.User, 'f', 1, 64),
		`"system":` + strconv.FormatFloat(c.System, 'f', 1, 64),
		`"idle":` + strconv.FormatFloat(c.Idle, 'f', 1, 64),
		`"nice":` + strconv.FormatFloat(c.Nice, 'f', 1, 64),
		`"iowait":` + strconv.FormatFloat(c.Iowait, 'f', 1, 64),
		`"irq":` + strconv.FormatFloat(c.Irq, 'f', 1, 64),
		`"softirq":` + strconv.FormatFloat(c.Softirq, 'f', 1, 64),
		`"steal":` + strconv.FormatFloat(c.Steal, 'f', 1, 64),
		`"guest":` + strconv.FormatFloat(c.Guest, 'f', 1, 64),
		`"guestNice":` + strconv.FormatFloat(c.GuestNice, 'f', 1, 64),
	}

	return `{` + strings.Join(v, ",") + `}`
}

// Deprecated: Total returns the total number of seconds in a CPUTimesStat
// Please do not use this internal function.
func (c TimesStat) Total() float64 {
	total := c.User + c.System + c.Idle + c.Nice + c.Iowait + c.Irq +
		c.Softirq + c.Steal + c.Guest + c.GuestNice

	return total
}

func (c InfoStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func getAllBusy(t TimesStat) (float64, float64) {
	tot := t.Total()
	if runtime.GOOS == "linux" {
		tot -= t.Guest     // Linux 2.6.24+
		tot -= t.GuestNice // Linux 3.2.0+
	}

	busy := tot - t.Idle - t.Iowait

	return tot, busy
}

func calculateBusy(t1, t2 TimesStat) float64 {
	t1All, t1Busy := getAllBusy(t1)
	t2All, t2Busy := getAllBusy(t2)

	if t2Busy <= t1Busy {
		return 0
	}
	if t2All <= t1All {
		return 100
	}
	return math.Min(100, math.Max(0, (t2Busy-t1Busy)/(t2All-t1All)*100))
}

func calculateAllBusy(t1, t2 []TimesStat) ([]float64, error) {
	// Make sure the CPU measurements have the same length.
	if len(t1) != len(t2) {
		return nil, fmt.Errorf(
			"received two CPU counts: %d != %d",
			len(t1), len(t2),
		)
	}

	ret := make([]float64, len(t1))
	for i, t := range t2 {
		ret[i] = calculateBusy(t1[i], t)
	}
	return ret, nil
}

func calculateItem(v1, v2, duration float64) float64 {
	if v2 <= v1 {
		return 0
	}
	if duration <= 0 {
		return 0
	}

	return math.Min(100, math.Max(0, (v2-v1)/(duration)*100))
}

func calculateItems(t1, t2 TimesStat) TimesStat {
	duration := t2.Total() - t1.Total()

	items := TimesStat{
		CPU:       t1.CPU,
		User:      calculateItem(t1.User, t2.User, duration),
		System:    calculateItem(t1.System, t2.System, duration),
		Idle:      calculateItem(t1.Idle, t2.Idle, duration),
		Nice:      calculateItem(t1.Nice, t2.Nice, duration),
		Iowait:    calculateItem(t1.Iowait, t2.Iowait, duration),
		Irq:       calculateItem(t1.Irq, t2.Irq, duration),
		Softirq:   calculateItem(t1.Softirq, t2.Softirq, duration),
		Steal:     calculateItem(t1.Steal, t2.Steal, duration),
		Guest:     calculateItem(t1.Guest, t2.Guest, duration),
		GuestNice: calculateItem(t1.GuestNice, t2.GuestNice, duration),
	}

	return items
}

func calculateAllItems(t1, t2 []TimesStat) ([]TimesStat, error) {
	// Make sure the CPU measurements have the same length.
	if len(t1) != len(t2) {
		return nil, fmt.Errorf(
			"received two CPU counts: %d != %d",
			len(t1), len(t2),
		)
	}

	ret := make([]TimesStat, len(t1))
	for i, t := range t2 {
		if t1[i].CPU != t.CPU {
			return nil, fmt.Errorf(
				"CPU number mismatch at %d: %s != %s",
				i, t1[i].CPU, t.CPU,
			)
		}
		ret[i] = calculateItems(t1[i], t)
	}
	return ret, nil
}

// Percent calculates the percentage of cpu used either per CPU or combined.
// If an interval of 0 is given it will compare the current cpu times against the last call.
// Returns one value per cpu, or a single value if percpu is set to false.
func Percent(interval time.Duration, percpu bool) ([]float64, error) {
	return PercentWithContext(context.Background(), interval, percpu)
}

// CPUTimesPercent calculates the percentage of CPU time used for different type of work
// either per CPU or combined.
// If an interval of 0 is given it will compare the current cpu times against the last call.
// Returns one value per cpu, or a single value if percpu is set to false.
// When interval is too small, returned values can be all zero.
func CPUTimesPercent(interval time.Duration, percpu bool) ([]TimesStat, error) {
	return CPUTimesPercentWithContext(context.Background(), interval, percpu)
}

func PercentWithContext(ctx context.Context, interval time.Duration, percpu bool) ([]float64, error) {
	t1, t2, err := sample(ctx, interval, percpu)
	if err != nil {
		return nil, err
	}

	return calculateAllBusy(t1, t2)
}

func CPUTimesPercentWithContext(ctx context.Context, interval time.Duration, percpu bool) ([]TimesStat, error) {
	t1, t2, err := sample(ctx, interval, percpu)
	if err != nil {
		return nil, err
	}

	return calculateAllItems(t1, t2)
}

func sample(ctx context.Context, interval time.Duration, percpu bool) (cpuTimes1, cpuTimes2 []TimesStat, err error) {
	if interval <= 0 {
		return sampleAndSaveAsLastWithContext(ctx, percpu)
	}

	// Get CPU usage at the start of the interval.
	cpuTimes1, err = TimesWithContext(ctx, percpu)
	if err != nil {
		return
	}

	if err := common.Sleep(ctx, interval); err != nil {
		return nil, nil, err
	}

	// And at the end of the interval.
	cpuTimes2, err = TimesWithContext(ctx, percpu)
	if err != nil {
		return
	}

	return
}

func sampleAndSaveAsLast(percpu bool) ([]TimesStat, []TimesStat, error) {
	return sampleAndSaveAsLastWithContext(context.Background(), percpu)
}

func sampleAndSaveAsLastWithContext(ctx context.Context, percpu bool) ([]TimesStat, []TimesStat, error) {
	cpuTimes, err := TimesWithContext(ctx, percpu)
	if err != nil {
		return nil, nil, err
	}
	lastCPUPercent.Lock()
	defer lastCPUPercent.Unlock()
	var lastTimes []TimesStat
	if percpu {
		lastTimes = lastCPUPercent.lastPerCPUTimes
		lastCPUPercent.lastPerCPUTimes = cpuTimes
	} else {
		lastTimes = lastCPUPercent.lastCPUTimes
		lastCPUPercent.lastCPUTimes = cpuTimes
	}

	if lastTimes == nil {
		return nil, nil, fmt.Errorf("error getting times for cpu percent. lastTimes was nil")
	}
	return lastTimes, cpuTimes, nil
}
