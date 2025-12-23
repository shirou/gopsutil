// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package load

import (
	"context"
)

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

// Misc returns miscellaneous host-wide statistics.
// darwin use ps command to get process running/blocked count.
// Almost same as Darwin implementation, but state is different.
func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

// SystemCalls returns the number of system calls since boot.
func SystemCalls() (int, error) {
	return SystemCallsWithContext(context.Background())
}

// Interrupts returns the number of interrupts since boot.
func Interrupts() (int, error) {
	return InterruptsWithContext(context.Background())
}
