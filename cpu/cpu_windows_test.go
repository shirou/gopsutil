// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package cpu

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// timesTotalSlack is the per-logical-CPU margin added on top of the measured
// wall clock time when comparing cpu-total against the sum of the per-CPU stats.
// It only has to absorb the counter granularity, since the time the two Times
// calls are actually apart is measured rather than assumed.
const timesTotalSlack = 0.1 // seconds per logical CPU

// TestPerfInfoMatchesLogicalCount ensures perfInfo() returns one entry per logical
// CPU on the host. This guards against regressions like issue #887 where only the
// calling thread's processor group was reported on hosts with more than 64 CPUs.
func TestPerfInfoMatchesLogicalCount(t *testing.T) {
	info, err := perfInfo()
	require.NoError(t, err)

	n, err := CountsWithContext(context.Background(), true)
	require.NoError(t, err)

	require.Len(t, info, n, "perfInfo must return one entry per logical CPU across all processor groups")
}

// TestTimesTotalMatchesPerCPUSum ensures the cpu-total entry is the field-wise sum
// of the per-CPU entries. Both are derived from the same perfInfo() counters, so
// they must agree apart from the counters advancing between the two calls. It
// guards the accumulation loop in TimesWithContext: a mis-mapped field, a dropped
// counter or a skipped processor group shows up as a mismatch.
//
// Note this does not by itself detect a revert to a GetSystemTimes-based total:
// that call reports the sum over the calling thread's processor group, which is
// the whole machine on a single-group host, so User/System/Idle would still match.
// Only the Irq assertion would catch it, and only once enough interrupt time has
// accumulated, since GetSystemTimes cannot report it at all.
func TestTimesTotalMatchesPerCPUSum(t *testing.T) {
	start := time.Now()

	perCPU, err := Times(true)
	require.NoError(t, err)
	require.NotEmpty(t, perCPU)

	total, err := Times(false)
	require.NoError(t, err)
	elapsed := time.Since(start)
	require.Len(t, total, 1)
	require.Equal(t, "cpu-total", total[0].CPU)

	var want TimesStat
	for _, c := range perCPU {
		want.User += c.User
		want.System += c.System
		want.Idle += c.Idle
		want.Irq += c.Irq
	}

	// Each logical CPU can advance by at most the wall clock time between the two
	// calls, so scale the tolerance by what actually elapsed. Assuming the calls
	// are close together would make this flaky whenever the test goroutine is
	// descheduled on a busy CI host.
	slack := (elapsed.Seconds() + timesTotalSlack) * float64(len(perCPU))
	assert.InDelta(t, want.User, total[0].User, slack)
	assert.InDelta(t, want.System, total[0].System, slack)
	assert.InDelta(t, want.Idle, total[0].Idle, slack)
	assert.InDelta(t, want.Irq, total[0].Irq, slack)
}

// TestSystemTimes exercises the GetSystemTimes fallback directly. TimesWithContext
// only reaches it when perfInfo() fails, which never happens on a healthy host, so
// without this test the fallback would ship untested.
func TestSystemTimes(t *testing.T) {
	total, err := systemTimes()
	require.NoError(t, err)
	require.Len(t, total, 1)

	assert.Equal(t, "cpu-total", total[0].CPU)
	assert.Positive(t, total[0].User)
	assert.Positive(t, total[0].Idle)
	assert.GreaterOrEqual(t, total[0].System, 0.0)
	// GetSystemTimes reports no interrupt time, so Irq stays unset here. This is
	// the one field that differs from the perfInfo-based total.
	assert.Zero(t, total[0].Irq)

	// The counters are cumulative since boot, so a second call must not go
	// backwards. This catches a broken FILETIME recombination, which would
	// otherwise only show up as an implausible absolute value.
	again, err := systemTimes()
	require.NoError(t, err)
	require.Len(t, again, 1)
	assert.GreaterOrEqual(t, again[0].User, total[0].User)
	assert.GreaterOrEqual(t, again[0].System, total[0].System)
	assert.GreaterOrEqual(t, again[0].Idle, total[0].Idle)
}
