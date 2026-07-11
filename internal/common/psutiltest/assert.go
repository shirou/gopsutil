// SPDX-License-Identifier: BSD-3-Clause
package psutiltest

import (
	"cmp"
	"fmt"
	"math"
	"testing"
	"time"
)

// DefaultTimeout and DefaultTick are the recommended budgets for
// require.EventuallyWithT when resampling fluctuating gauges; the
// timeout spans at least one loadavg update interval (5 seconds).
const (
	DefaultTimeout = 10 * time.Second
	DefaultTick    = 2 * time.Second
)

// AssertBracketed asserts before <= got <= after. Use it for monotonic
// counters: sample psutil before and after the gopsutil call and the
// gopsutil value must fall inside the bracket.
func AssertBracketed[T cmp.Ordered](tb testing.TB, field string, before, got, after T) {
	tb.Helper()
	if got < before || got > after {
		tb.Errorf("%s: gopsutil value %v not in psutil bracket [%v, %v]", field, got, before, after)
	}
}

// AssertBracketedDelta is the float64 form of AssertBracketed with
// slack for rounding differences: before-slack <= got <= after+slack.
func AssertBracketedDelta(tb testing.TB, field string, before, got, after, slack float64) {
	tb.Helper()
	if got < before-slack || got > after+slack {
		tb.Errorf("%s: gopsutil value %v not in psutil bracket [%v, %v] (slack %v)", field, got, before, after, slack)
	}
}

// CheckWithinTolerance returns an error unless
// |expected-actual| <= max(rel*|expected|, absFloor). The absolute
// floor keeps small or zero expected values (e.g. Buffers) from turning
// a purely relative comparison into a flake. The error form suits
// assert.NoError inside require.EventuallyWithT callbacks.
func CheckWithinTolerance(field string, expected, actual, rel, absFloor float64) error {
	tol := math.Max(rel*math.Abs(expected), absFloor)
	if diff := math.Abs(expected - actual); diff > tol {
		return fmt.Errorf("%s: psutil %v vs gopsutil %v differ by %v (tolerance %v)", field, expected, actual, diff, tol)
	}
	return nil
}
