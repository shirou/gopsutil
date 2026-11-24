// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package disk

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// On APFS, statfs Blocks-Bfree gives the used space of the whole container,
// not of this volume, so Used must come from the volume itself, as df -k
// reports it.
func TestUsageDarwinUsedMatchesDf(t *testing.T) {
	v, err := UsageWithContext(context.Background(), "/")
	require.NoError(t, err)

	out, err := invoke.Command("/bin/df", "-k", "/")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	require.Len(t, lines, 2)
	fields := strings.Fields(lines[1])
	require.GreaterOrEqual(t, len(fields), 4)
	dfUsed, err := strconv.ParseUint(fields[2], 10, 64)
	require.NoError(t, err)

	// df and disk.Usage are sampled at different times; allow 1% of Total
	// for writes that happen in between.
	dfUsed *= 1024
	diff := v.Used - dfUsed
	if dfUsed > v.Used {
		diff = dfUsed - v.Used
	}
	assert.LessOrEqualf(t, diff, v.Total/100, "Used=%d df used=%d total=%d", v.Used, dfUsed, v.Total)
}
