//go:build openbsd
// +build openbsd

package cpu

import (
	"context"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/shirou/gopsutil/v3/internal/common"
	"github.com/tklauser/go-sysconf"
	"golang.org/x/sys/unix"
)

import "C"

const (
	// sys/sched.h
	cpUser    = 0
	cpNice    = 1
	cpSys     = 2
	cpSpin    = 3
	cpIntr    = 4
	cpIdle    = 5
	cpuStates = 6
	cpuOnline = 0x0001 // CPUSTATS_ONLINE

	// sys/sysctl.h
	ctlKern      = 1  // "high kernel": proc, limits
	ctlHw        = 6  // CTL_HW
	smt          = 24 // HW_SMT
	kernCpTime   = 40 // KERN_CPTIME
	kernCPUStats = 85 // KERN_CPUSTATS
)

var ClocksPerSec = float64(128)

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

func cpsToTS(cpuTimes [cpuStates]uint64, name string) TimesStat {
	return TimesStat{
		CPU:    name,
		User:   float64(cpuTimes[cpUser]) / ClocksPerSec,
		Nice:   float64(cpuTimes[cpNice]) / ClocksPerSec,
		System: float64(cpuTimes[cpSys]) / ClocksPerSec,
		Idle:   float64(cpuTimes[cpIdle]) / ClocksPerSec,
		Irq:    float64(cpuTimes[cpIntr]) / ClocksPerSec,
	}
}

func TimesWithContext(ctx context.Context, percpu bool) (ret []TimesStat, err error) {
	cpuTimes := [cpuStates]uint64{}

	if !percpu {
		mib := []int32{ctlKern, kernCpTime}
		buf, _, err := common.CallSyscall(mib)
		if err != nil {
			return ret, err
		}
		var x []C.long
		// could use unsafe.Slice but it's only for go1.17+
		x = (*[cpuStates]C.long)(unsafe.Pointer(&buf[0]))[:]
		for i := range x {
			cpuTimes[i] = uint64(x[i])
		}
		c := cpsToTS(cpuTimes, "cpu-total")
		return []TimesStat{c}, nil
	}

	ncpu, err := unix.SysctlUint32("hw.ncpu")
	if err != nil {
		return
	}

	var i uint32
	for i = 0; i < ncpu; i++ {
		mib := []int32{ctlKern, kernCPUStats, int32(i)}
		buf, _, err := common.CallSyscall(mib)
		if err != nil {
			return ret, err
		}

		data := unsafe.Pointer(&buf[0])
		fptr := unsafe.Pointer(uintptr(data) + uintptr(8*cpuStates))
		flags := *(*uint64)(fptr)
		if (flags & cpuOnline) == 0 {
			continue
		}

		var x []uint64
		x = (*[cpuStates]uint64)(data)[:]
		for i := range x {
			cpuTimes[i] = x[i]
		}
		c := cpsToTS(cpuTimes, fmt.Sprintf("cpu%d", i))
		ret = append(ret, c)
	}

	return ret, nil
}

// Returns only one (minimal) CPUInfoStat on OpenBSD
func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
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

func CountsWithContext(ctx context.Context, logical bool) (int, error) {
	return runtime.NumCPU(), nil
}
