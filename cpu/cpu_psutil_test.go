// SPDX-License-Identifier: BSD-3-Clause
package cpu

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
	cpuTimesSlack = 0.2 // seconds; rounding differences between samples
	iowaitDelta   = 1.0 // seconds; iowait can decrease, see proc(5)
)

// Percent and Info are deliberately not compared: Percent is an
// instantaneous value and Info has no direct psutil equivalent.

func TestCounts_Against_Psutil(t *testing.T) {
	pt.RequirePsutil(t)

	// The linux implementation is a direct port of psutil's algorithm
	// (see the links in cpu_linux.go), so the counts must match exactly.
	for _, tc := range []struct {
		name    string
		logical bool
		expr    string
	}{
		{"logical", true, "psutil.cpu_count(logical=True)"},
		{"physical", false, "psutil.cpu_count(logical=False)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Counts(tc.logical)
			if errors.Is(err, common.ErrNotImplementedError) {
				t.Skip("not implemented")
			}
			require.NoError(t, err)

			want, err := pt.Eval[*int](t, tc.expr)
			require.NoError(t, err)
			if want == nil {
				// psutil returns None for a count it cannot determine
				// (e.g. some containers).
				t.Skip("psutil could not determine the count")
			}
			assert.Equal(t, *want, got)
		})
	}
}

type pyCPUTimes struct {
	User    float64 `json:"user"`
	System  float64 `json:"system"`
	Idle    float64 `json:"idle"`
	Nice    float64 `json:"nice"`
	Iowait  float64 `json:"iowait"`
	Irq     float64 `json:"irq"`
	Softirq float64 `json:"softirq"`
	Steal   float64 `json:"steal"`
}

func TestTimes_Against_Psutil(t *testing.T) {
	pt.RequirePsutil(t)

	before, err := pt.Eval[pyCPUTimes](t, "psutil.cpu_times(percpu=False)")
	require.NoError(t, err)

	times, err := Times(false)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	require.Len(t, times, 1)
	got := times[0]

	after, err := pt.Eval[pyCPUTimes](t, "psutil.cpu_times(percpu=False)")
	require.NoError(t, err)

	// user/system/idle exist on all supported platforms and both sides
	// derive them identically (linux: /proc/stat; windows: GetSystemTimes
	// with system = kernel - idle; darwin: mach host statistics).
	pt.AssertBracketedDelta(t, "User", before.User, got.User, after.User, cpuTimesSlack)
	pt.AssertBracketedDelta(t, "System", before.System, got.System, after.System, cpuTimesSlack)
	pt.AssertBracketedDelta(t, "Idle", before.Idle, got.Idle, after.Idle, cpuTimesSlack)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// nice is also common to linux (/proc/stat) and darwin, where both
		// sides read the same mach nice tick counter.
		pt.AssertBracketedDelta(t, "Nice", before.Nice, got.Nice, after.Nice, cpuTimesSlack)
	}

	if runtime.GOOS == "linux" {
		// The remaining fields exist only in linux's /proc/stat accounting,
		// which both sides report raw, in seconds.
		pt.AssertBracketedDelta(t, "Irq", before.Irq, got.Irq, after.Irq, cpuTimesSlack)
		pt.AssertBracketedDelta(t, "Softirq", before.Softirq, got.Softirq, after.Softirq, cpuTimesSlack)
		pt.AssertBracketedDelta(t, "Steal", before.Steal, got.Steal, after.Steal, cpuTimesSlack)
		// iowait can decrease (see proc(5)) and, as a sum over all CPUs,
		// can legitimately advance by more than the slack between two
		// samples on busy many-core hosts, so bracket it between the
		// smaller and larger psutil sample with slack on both ends.
		lo, hi := min(before.Iowait, after.Iowait), max(before.Iowait, after.Iowait)
		pt.AssertBracketedDelta(t, "Iowait", lo, got.Iowait, hi, iowaitDelta)
	}
}

func TestTimesPerCPU_Against_Psutil(t *testing.T) {
	pt.RequirePsutil(t)

	before, err := pt.Eval[[]pyCPUTimes](t, "psutil.cpu_times(percpu=True)")
	require.NoError(t, err)

	times, err := Times(true)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	after, err := pt.Eval[[]pyCPUTimes](t, "psutil.cpu_times(percpu=True)")
	require.NoError(t, err)

	require.Len(t, times, len(before))
	require.Len(t, after, len(before))

	// Both sides enumerate CPUs in the same order (/proc/stat order on linux).
	for i, got := range times {
		pt.AssertBracketedDelta(t, got.CPU+".User", before[i].User, got.User, after[i].User, cpuTimesSlack)
		pt.AssertBracketedDelta(t, got.CPU+".System", before[i].System, got.System, after[i].System, cpuTimesSlack)
		pt.AssertBracketedDelta(t, got.CPU+".Idle", before[i].Idle, got.Idle, after[i].Idle, cpuTimesSlack)
	}
}
