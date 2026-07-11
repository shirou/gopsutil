// SPDX-License-Identifier: BSD-3-Clause
package mem

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
	pt "github.com/shirou/gopsutil/v4/internal/common/psutiltest"
)

const (
	memRelTol    = 0.10
	memAbsFloor  = 64 << 20 // 64 MiB
	swapRelTol   = 0.05
	swapAbsFloor = 16 << 20 // 16 MiB
)

func TestVirtualMemory_Against_Psutil(t *testing.T) {
	pt.RequirePsutil(t)

	type pyVirtualMemory struct {
		Total     uint64 `json:"total"`
		Available uint64 `json:"available"`
		Free      uint64 `json:"free"`
		Buffers   uint64 `json:"buffers"`
		Cached    uint64 `json:"cached"`
		Shared    uint64 `json:"shared"`
		Slab      uint64 `json:"slab"`
	}

	v, err := VirtualMemory()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	py, err := pt.Eval[pyVirtualMemory](t, "psutil.virtual_memory()")
	require.NoError(t, err)

	// Total is static and read from the same source on every OS
	// (linux: MemTotal, darwin: hw.memsize, windows: GlobalMemoryStatusEx).
	require.Equal(t, py.Total, v.Total)

	// Used and UsedPercent are deliberately not compared: gopsutil defines
	// Used = Total - Available while psutil uses total - free - buffers - cached.

	// The remaining fields fluctuate, so resample both sides until they agree.
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		v, err := VirtualMemory()
		if !assert.NoError(c, err) {
			return
		}
		py, err := pt.Eval[pyVirtualMemory](t, "psutil.virtual_memory()")
		if !assert.NoError(c, err) {
			return
		}
		// Available comes from the same source on every OS (linux:
		// MemAvailable, darwin: free+inactive, windows: AvailPhys).
		assert.NoError(c, pt.CheckWithinTolerance("Available", float64(py.Available), float64(v.Available), memRelTol, memAbsFloor))
		if runtime.GOOS != "linux" {
			// The fields below are only compared on linux. Free is known to
			// differ on darwin: psutil subtracts speculative pages from the
			// mach free count while gopsutil reports it raw.
			return
		}
		// Both sides read /proc/meminfo and both fold SReclaimable into Cached.
		assert.NoError(c, pt.CheckWithinTolerance("Free", float64(py.Free), float64(v.Free), memRelTol, memAbsFloor))
		assert.NoError(c, pt.CheckWithinTolerance("Buffers", float64(py.Buffers), float64(v.Buffers), memRelTol, memAbsFloor))
		assert.NoError(c, pt.CheckWithinTolerance("Cached", float64(py.Cached), float64(v.Cached), memRelTol, memAbsFloor))
		assert.NoError(c, pt.CheckWithinTolerance("Shared", float64(py.Shared), float64(v.Shared), memRelTol, memAbsFloor))
		assert.NoError(c, pt.CheckWithinTolerance("Slab", float64(py.Slab), float64(v.Slab), memRelTol, memAbsFloor))
	}, pt.DefaultTimeout, pt.DefaultTick)
}

func TestSwapMemory_Against_Psutil(t *testing.T) {
	if runtime.GOOS != "linux" {
		// psutil's swap semantics on windows (pagefile stats) and darwin
		// (sysctl vm.swapusage granularity) have not been verified to match
		// gopsutil's yet.
		t.Skip("swap comparison is only verified on linux")
	}
	pt.RequirePsutil(t)

	type pySwapMemory struct {
		Total uint64 `json:"total"`
		Used  uint64 `json:"used"`
		Free  uint64 `json:"free"`
		Sin   uint64 `json:"sin"`
		Sout  uint64 `json:"sout"`
	}

	before, err := pt.Eval[pySwapMemory](t, "psutil.swap_memory()")
	require.NoError(t, err)

	v, err := SwapMemory()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	require.Equal(t, before.Total, v.Total)

	if v.Total == 0 {
		t.Skip("no swap configured")
	}

	// Sin/Sout: both sides multiply pswpin/pswpout from /proc/vmstat by a
	// hardcoded 4096 (gopsutil mem_linux.go, psutil _pslinux.py), so the
	// values are directly comparable regardless of the actual page size.
	after, err := pt.Eval[pySwapMemory](t, "psutil.swap_memory()")
	require.NoError(t, err)
	pt.AssertBracketed(t, "Sin", before.Sin, v.Sin, after.Sin)
	pt.AssertBracketed(t, "Sout", before.Sout, v.Sout, after.Sout)

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		v, err := SwapMemory()
		if !assert.NoError(c, err) {
			return
		}
		py, err := pt.Eval[pySwapMemory](t, "psutil.swap_memory()")
		if !assert.NoError(c, err) {
			return
		}
		assert.NoError(c, pt.CheckWithinTolerance("Used", float64(py.Used), float64(v.Used), swapRelTol, swapAbsFloor))
		assert.NoError(c, pt.CheckWithinTolerance("Free", float64(py.Free), float64(v.Free), swapRelTol, swapAbsFloor))
	}, pt.DefaultTimeout, pt.DefaultTick)
}
