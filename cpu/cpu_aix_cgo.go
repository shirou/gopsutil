// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package cpu

import (
	"context"
	"time"

	"github.com/power-devops/perfstat"
	"github.com/shirou/gopsutil/v4/internal/common"
)

func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	var ret []TimesStat
	if percpu {
		cpus, err := perfstat.CpuStat()
		if err != nil {
			return nil, err
		}
		for _, c := range cpus {
			ct := &TimesStat{
				CPU:    c.Name,
				Idle:   float64(c.Idle),
				User:   float64(c.User),
				System: float64(c.Sys),
				Iowait: float64(c.Wait),
			}
			ret = append(ret, *ct)
		}
	} else {
		c, err := perfstat.CpuTotalStat()
		if err != nil {
			return nil, err
		}
		ct := &TimesStat{
			CPU:    "cpu-total",
			Idle:   float64(c.Idle),
			User:   float64(c.User),
			System: float64(c.Sys),
			Iowait: float64(c.Wait),
		}
		ret = append(ret, *ct)
	}
	return ret, nil
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	c, err := perfstat.CpuTotalStat()
	if err != nil {
		return nil, err
	}
	p, err := perfstat.LparInfo()
	if err != nil {
		return nil, err
	}
	info := InfoStat{
		CPU:       0,
		ModelName: c.Description,
		Mhz:       float64(c.ProcessorHz / 1000000),
		Cores:     int32(p.OnlineVCpus),
	}
	result := []InfoStat{info}
	return result, nil
}

func CountsWithContext(ctx context.Context, logical bool) (int, error) {
	if logical {
		c, err := perfstat.CpuTotalStat()
		if err != nil {
			return 0, err
		}
		return c.NCpusCfg, nil
	}
	// For physical count, use the number of online virtual CPUs (before SMT multiplications).
	p, err := perfstat.LparInfo()
	if err != nil {
		return 0, err
	}
	return int(p.OnlineVCpus), nil
}

// aixPercent returns ErrNotImplementedError for CGO builds because the CGO
// TimesWithContext returns cumulative tick counters (from perfstat), not
// instantaneous percentages. The caller in cpu.go falls through to the
// standard delta-based calculation when this returns an error.
func aixPercent(_ context.Context, _ time.Duration, _ bool) ([]float64, error) {
	return nil, common.ErrNotImplementedError
}
