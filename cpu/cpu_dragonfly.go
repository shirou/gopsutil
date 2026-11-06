// SPDX-License-Identifier: BSD-3-Clause
package cpu

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// cpuTimesFields is the number of counters at the start of struct kinfo_cputime.
// The struct has more fields, so its size must not be derived from these.
const cpuTimesFields = 5

// ClocksPerSec converts kern.cp_time, which DragonFly counts in microseconds.
// sysconf(_SC_CLK_TCK) does not apply to these counters.
var ClocksPerSec = float64(1000000)

func timeStat(name string, record []byte) TimesStat {
	field := func(i int) float64 {
		return float64(binary.LittleEndian.Uint64(record[i*8:])) / ClocksPerSec
	}
	return TimesStat{
		CPU:    name,
		User:   field(0),
		Nice:   field(1),
		System: field(2),
		Irq:    field(3),
		Idle:   field(4),
	}
}

func Times(percpu bool) ([]TimesStat, error) {
	return TimesWithContext(context.Background(), percpu)
}

func TimesWithContext(_ context.Context, percpu bool) ([]TimesStat, error) {
	if !percpu {
		buf, err := unix.SysctlRaw("kern.cp_time")
		if err != nil {
			return nil, err
		}
		if len(buf) < cpuTimesFields*8 {
			return nil, errors.New("kern.cp_time is too short")
		}
		return []TimesStat{timeStat("cpu-total", buf)}, nil
	}

	buf, err := unix.SysctlRaw("kern.cp_times")
	if err != nil {
		return nil, err
	}
	ncpu, err := unix.SysctlUint32("hw.ncpu")
	if err != nil {
		return nil, err
	}
	stride := len(buf) / int(ncpu)
	if stride < cpuTimesFields*8 {
		return nil, errors.New("kern.cp_times is too short")
	}
	ret := make([]TimesStat, 0, ncpu)
	for i := 0; i < int(ncpu); i++ {
		ret = append(ret, timeStat(fmt.Sprintf("cpu%d", i), buf[i*stride:]))
	}
	return ret, nil
}

// Info returns a single InfoStat: hw.cpu_topology.tree has no CHIP entries on
// virtual machines, so the number of packages cannot be derived reliably.
func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(_ context.Context) ([]InfoStat, error) {
	model, err := unix.Sysctl("hw.model")
	if err != nil {
		return nil, err
	}
	info := InfoStat{
		ModelName: strings.TrimSpace(model),
		Cores:     int32(runtime.NumCPU()),
	}
	if mhz, err := unix.SysctlUint32("hw.clockrate"); err == nil {
		info.Mhz = float64(mhz)
	}
	return []InfoStat{info}, nil
}

func CountsWithContext(_ context.Context, _ bool) (int, error) {
	return runtime.NumCPU(), nil
}
