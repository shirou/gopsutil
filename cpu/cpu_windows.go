// +build windows

package cpu

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/internal/common"
	"golang.org/x/sys/windows"
)

type Win32_Processor struct {
	LoadPercentage            *uint16
	Family                    uint16
	Manufacturer              string
	Name                      string
	NumberOfLogicalProcessors uint32
	ProcessorID               *string
	Stepping                  *string
	MaxClockSpeed             uint32
}

type win32_PerfRawData_Counters_ProcessorInformation struct {
	Name                  string
	PercentDPCTime        uint64
	PercentIdleTime       uint64
	PercentUserTime       uint64
	PercentProcessorTime  uint64
	PercentInterruptTime  uint64
	PercentPriorityTime   uint64
	PercentPrivilegedTime uint64
	InterruptsPerSec      uint32
	ProcessorFrequency    uint32
	DPCRate               uint32
}

type systemTimes struct {
	idleTime   uint64
	kernelTime uint64
	userTime   uint64
}

var lastSystemTimes systemTimes

func init() {
	totalCPUTimes()
}

// Win32_PerfFormattedData_PerfOS_System struct to have count of processes and processor queue length
type Win32_PerfFormattedData_PerfOS_System struct {
	Processes            uint32
	ProcessorQueueLength uint32
}

const (
	win32_TicksPerSecond = 10000000.0
)

// Times returns times stat per cpu and combined for all CPUs
func Times(percpu bool) ([]TimesStat, error) {
	return TimesWithContext(context.Background(), percpu)
}

func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	if percpu {
		return perCPUTimesWithContext(ctx)
	}

	return totalCPUTimes()
}

func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	var ret []InfoStat
	var dst []Win32_Processor
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithContext(ctx, q, &dst); err != nil {
		return ret, err
	}

	var procID string
	for i, l := range dst {
		procID = ""
		if l.ProcessorID != nil {
			procID = *l.ProcessorID
		}

		cpu := InfoStat{
			CPU:        int32(i),
			Family:     fmt.Sprintf("%d", l.Family),
			VendorID:   l.Manufacturer,
			ModelName:  l.Name,
			Cores:      int32(l.NumberOfLogicalProcessors),
			PhysicalID: procID,
			Mhz:        float64(l.MaxClockSpeed),
			Flags:      []string{},
		}
		ret = append(ret, cpu)
	}

	return ret, nil
}

// PerfInfo returns the performance counter's instance value for ProcessorInformation.
// Name property is the key by which overall, per cpu and per core metric is known.
func perfInfoWithContext(ctx context.Context) ([]win32_PerfRawData_Counters_ProcessorInformation, error) {
	var ret []win32_PerfRawData_Counters_ProcessorInformation

	q := wmi.CreateQuery(&ret, "WHERE NOT Name LIKE '%_Total'")
	err := common.WMIQueryWithContext(ctx, q, &ret)
	if err != nil {
		return []win32_PerfRawData_Counters_ProcessorInformation{}, err
	}

	return ret, err
}

// ProcInfo returns processes count and processor queue length in the system.
// There is a single queue for processor even on multiprocessors systems.
func ProcInfo() ([]Win32_PerfFormattedData_PerfOS_System, error) {
	return ProcInfoWithContext(context.Background())
}

func ProcInfoWithContext(ctx context.Context) ([]Win32_PerfFormattedData_PerfOS_System, error) {
	var ret []Win32_PerfFormattedData_PerfOS_System
	q := wmi.CreateQuery(&ret, "")
	err := common.WMIQueryWithContext(ctx, q, &ret)
	if err != nil {
		return []Win32_PerfFormattedData_PerfOS_System{}, err
	}
	return ret, err
}

// perCPUTimes returns times stat per cpu, per core and overall for all CPUs
func perCPUTimesWithContext(ctx context.Context) ([]TimesStat, error) {
	var ret []TimesStat
	stats, err := perfInfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range stats {
		c := TimesStat{
			CPU:    v.Name,
			User:   float64(v.PercentUserTime) / win32_TicksPerSecond,
			System: float64(v.PercentPrivilegedTime) / win32_TicksPerSecond,
			Idle:   float64(v.PercentIdleTime) / win32_TicksPerSecond,
			Irq:    float64(v.PercentInterruptTime) / win32_TicksPerSecond,
		}
		ret = append(ret, c)
	}
	return ret, nil
}

func totalCPUTimes() ([]TimesStat, error) {
	var ret []TimesStat
	currSystemTimes, err := getSystemTimes()

	if err != nil {
		return ret, err
	}

	if lastSystemTimes.idleTime != 0 {
		deltaTimes := systemTimes{
			idleTime:   currSystemTimes.idleTime - lastSystemTimes.idleTime,
			kernelTime: currSystemTimes.kernelTime - lastSystemTimes.kernelTime,
			userTime:   currSystemTimes.userTime - lastSystemTimes.userTime,
		}

		// Duration between consecutive API calls was enough to calculate CPU
		if (deltaTimes.userTime != 0 && deltaTimes.kernelTime != 0) {
			// kernelTime already contains idleTime
			deltaAll := deltaTimes.kernelTime + deltaTimes.userTime

			currUtilizationPerc := (1.0 - float64(deltaTimes.idleTime)/float64(deltaAll)) * 100
			currUserPerc := (float64(deltaTimes.userTime) / float64(deltaAll)) * 100

			ret = append(ret, TimesStat{
				CPU:    "cpu-total",
				User:   currUserPerc,
				System: 100.0 - (100.0 - currUtilizationPerc) - currUserPerc,
				Idle:   100.0 - currUtilizationPerc,
			})
		}
	}

	lastSystemTimes = currSystemTimes
	return ret, nil
}

func getSystemTimes() (systemTimes, error) {
	var lpIdleTime, lpKernelTime, lpUserTime common.FILETIME

	r, _, _ := common.ProcGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)))

	if r == 0 {
		return systemTimes{}, windows.GetLastError()
	}

	ret := systemTimes{
		idleTime:   convertFiletimeToInt64(lpIdleTime),
		kernelTime: convertFiletimeToInt64(lpKernelTime),
		userTime:   convertFiletimeToInt64(lpUserTime),
	}

	return ret, nil
}

func convertFiletimeToInt64(ft common.FILETIME) uint64 {
	return (uint64(ft.DwHighDateTime) << 32) | uint64(ft.DwLowDateTime)
}
