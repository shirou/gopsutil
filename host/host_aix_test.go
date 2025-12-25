// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootTimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.BootTimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	bootTime, err := BootTimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, bootTime)
}

func TestUptimeWithContext(t *testing.T) {
	// This is a wrapper function that delegates to common.UptimeWithContext
	// Actual implementation testing is done in common_aix_test.go
	uptime, err := UptimeWithContext(context.TODO())
	require.NoError(t, err)
	assert.Positive(t, uptime)
}

func TestFDLimitsWithContext(t *testing.T) {
	ctx := context.Background()
	soft, hard, err := FDLimitsWithContext(ctx)
	require.NoError(t, err)

	// Both limits should be positive
	assert.Positive(t, soft, "Soft limit should be > 0")
	assert.Positive(t, hard, "Hard limit should be > 0")

	// Hard limit should be >= soft limit
	assert.GreaterOrEqual(t, hard, soft, "Hard limit should be >= soft limit")

	// Reasonable ranges for AIX (typically 1024-32767, or unlimited which is max int64)
	assert.GreaterOrEqual(t, soft, uint64(256), "Soft limit should be >= 256")
}

func TestFDLimits(t *testing.T) {
	soft, hard, err := FDLimits()
	require.NoError(t, err)

	// Both limits should be positive
	assert.Positive(t, soft)
	assert.Positive(t, hard)
	assert.GreaterOrEqual(t, hard, soft)
}
