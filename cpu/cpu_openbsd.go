// SPDX-License-Identifier: BSD-3-Clause
//go:build openbsd

package cpu

import (
	"context"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/tklauser/go-sysconf"
	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

const (
	// sys/sched.h
	cpuOnline = 0x0001 // CPUSTATS_ONLINE

	// sys/sysctl.h
	ctlKern      = 1  // "high kernel": proc, limits
	ctlHw        = 6  // CTL_HW
	smt          = 24 // HW_SMT
	kernCPUStats = 85 // KERN_CPUSTATS
)

var ClocksPerSec = float64(128)

type cpuStats struct {
	// cs_time[CPUSTATES]
	User uint64
	Nice uint64
	Sys  uint64
	Spin uint64
	Intr uint64
	Idle uint64

	// cs_flags
	Flags uint64
}

func init() {
	clkTck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	// ignore errors
	if err == nil {
		ClocksPerSec = float64(clkTck)
	}
}

func Times(percpu bool) ([]TimesStat, error) {
	return TimesWithContext(context.Background(), percpu)
}

func TimesWithContext(_ context.Context, percpu bool) ([]TimesStat, error) {
	ret := make([]TimesStat, 0)

	// KERN_CPTIME is the per-CPU average (the kernel divides the sum by
	// hw.ncpuonline), so cpu-total would not match the sum of per-CPU
	// counters on multi-CPU hosts. Build the total from KERN_CPUSTATS
	// instead, which reports un-averaged per-CPU times.
	ncpu, err := unix.SysctlUint32("hw.ncpu")
	if err != nil {
		return ret, err
	}

	var i uint32
	for i = 0; i < ncpu; i++ {
		mib := []int32{ctlKern, kernCPUStats, int32(i)}
		buf, _, err := common.CallSyscall(mib)
		if err != nil {
			return ret, err
		}
		if len(buf) < int(unsafe.Sizeof(cpuStats{})) {
			continue
		}

		stats := (*cpuStats)(unsafe.Pointer(&buf[0]))
		if (stats.Flags & cpuOnline) == 0 {
			continue
		}
		ret = append(ret, TimesStat{
			CPU:    fmt.Sprintf("cpu%d", i),
			User:   float64(stats.User) / ClocksPerSec,
			Nice:   float64(stats.Nice) / ClocksPerSec,
			System: float64(stats.Sys) / ClocksPerSec,
			Idle:   float64(stats.Idle) / ClocksPerSec,
			Irq:    float64(stats.Intr) / ClocksPerSec,
		})
	}

	if !percpu {
		return []TimesStat{sumTimesStats(ret)}, nil
	}
	return ret, nil
}

// Returns only one (minimal) CPUInfoStat on OpenBSD
func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(_ context.Context) ([]InfoStat, error) {
	var ret []InfoStat
	var err error

	c := InfoStat{}

	mhz, err := unix.SysctlUint32("hw.cpuspeed")
	if err != nil {
		return nil, err
	}
	c.Mhz = float64(mhz)

	ncpu, err := unix.SysctlUint32("hw.ncpuonline")
	if err != nil {
		return nil, err
	}
	c.Cores = int32(ncpu)

	if c.ModelName, err = unix.Sysctl("hw.model"); err != nil {
		return nil, err
	}

	return append(ret, c), nil
}

func CountsWithContext(_ context.Context, _ bool) (int, error) {
	return runtime.NumCPU(), nil
}
