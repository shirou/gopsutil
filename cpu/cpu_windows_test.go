// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package cpu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

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
