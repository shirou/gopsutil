// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package cpu

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimesAIXUnits checks that AIX CPU times are reported in seconds, not raw clock ticks.
// Over a wall-clock interval, the summed CPU time per logical CPU cannot exceed the interval length.
func TestTimesAIXUnits(t *testing.T) {
	ctx := context.Background()

	require.Greater(t, ClocksPerSec, float64(0), "clock tick rate must be positive")

	before, err := TimesWithContext(ctx, true)
	require.NoError(t, err)
	start := time.Now()
	time.Sleep(2 * time.Second)
	after, err := TimesWithContext(ctx, true)
	require.NoError(t, err)
	elapsed := time.Since(start).Seconds()

	require.Equal(t, len(before), len(after), "CPU count changed between samples")
	require.NotEmpty(t, after)

	var total float64
	for i := range after {
		b, a := before[i], after[i]
		total += (a.User - b.User) + (a.System - b.System) + (a.Idle - b.Idle) + (a.Iowait - b.Iowait)
	}
	// Aggregate across CPUs. Raw clock ticks would make this about 100x the CPU count times wall clock.
	// Allow 2x the CPU count times wall clock to tolerate perfstat counter drift on shared LPARs.
	maxSeconds := float64(len(after)) * elapsed * 2
	assert.LessOrEqualf(t, total, maxSeconds, "total CPU time %.2fs over %.2fs wall clock across %d CPUs; values look like raw ticks", total, elapsed, len(after))
}

// TestTimesAIXClockTicks checks ClocksPerSec against the system's CLK_TCK, read via getconf.
func TestTimesAIXClockTicks(t *testing.T) {
	out, err := exec.Command("getconf", "CLK_TCK").Output()
	require.NoError(t, err, "getconf CLK_TCK failed")

	clkTck, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	require.NoError(t, err, "unexpected getconf output: %q", out)
	assert.Equal(t, clkTck, ClocksPerSec, "ClocksPerSec does not match getconf CLK_TCK")
}
