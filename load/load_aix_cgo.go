// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package load

/*
#cgo LDFLAGS: -L/usr/lib -lperfstat

#include <libperfstat.h>
#include <procinfo.h>
#include <sys/thread.h>
*/
import "C"

import (
	"context"
	"unsafe"

	"github.com/power-devops/perfstat"
	"github.com/shirou/gopsutil/v4/internal/common"
)

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	c, err := perfstat.CpuTotalStat()
	if err != nil {
		return nil, err
	}
	ret := &AvgStat{
		Load1:  float64(c.LoadAvg1),
		Load5:  float64(c.LoadAvg5),
		Load15: float64(c.LoadAvg15),
	}

	return ret, nil
}

func MiscWithContext(ctx context.Context) (*MiscStat, error) {
	// Count total processes and collect PIDs for thread-state enumeration.
	pinfo := C.struct_procentry64{}
	cpid := C.pid_t(0)

	ret := MiscStat{}
	var pids []C.pid_t
	for {
		num, err := C.getprocs64(unsafe.Pointer(&pinfo), C.sizeof_struct_procentry64, nil, 0, &cpid, 1)
		if err != nil {
			return nil, err
		}
		if num == 0 {
			break
		}
		ret.ProcsTotal++
		pids = append(pids, pinfo.pi_pid)
	}

	// Count threads in TSRUN state (runnable/running) across all processes.
	// SACTIVE at the process level means "active in memory" which includes
	// sleeping processes and is not a useful proxy for ProcsRunning.
	tinfo := C.struct_thrdentry64{}
	for _, pid := range pids {
		ctid := C.tid64_t(0)
		for {
			n, err := C.getthrds64(pid, unsafe.Pointer(&tinfo), C.sizeof_struct_thrdentry64, &ctid, 1)
			if err != nil {
				break // process may have exited
			}
			if n == 0 {
				break
			}
			if tinfo.ti_state == C.TSRUN {
				ret.ProcsRunning++
			}
		}
	}

	// ProcsBlocked: processes sleeping waiting for I/O, equivalent to Linux D state.
	// perfstat IOWait and PhysIO are instantaneous counts, matching vmstat's b column.
	var cpuStat C.perfstat_cpu_total_t
	if rc := C.perfstat_cpu_total(nil, &cpuStat, C.sizeof_perfstat_cpu_total_t, 1); rc == 1 {
		ret.ProcsBlocked = int(cpuStat.iowait) + int(cpuStat.physio)
	}

	return &ret, nil
}

func SystemCallsWithContext(_ context.Context) (int, error) {
	return 0, common.ErrNotImplementedError
}

func InterruptsWithContext(_ context.Context) (int, error) {
	return 0, common.ErrNotImplementedError
}
