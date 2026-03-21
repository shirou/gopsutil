// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package load

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiscWithContextAIX(t *testing.T) {
	ctx := context.Background()
	misc, err := MiscWithContext(ctx)
	require.NoError(t, err)
	assert.NotNil(t, misc)

	// Process counts should be non-negative
	assert.GreaterOrEqual(t, misc.ProcsTotal, 0)
	assert.GreaterOrEqual(t, misc.ProcsRunning, 0)
	assert.GreaterOrEqual(t, misc.ProcsBlocked, 0)

	// Total should be >= running + blocked
	assert.GreaterOrEqual(t, misc.ProcsTotal, misc.ProcsRunning+misc.ProcsBlocked)

	// Cumulative since-boot counters should be positive on a running system
	assert.Positive(t, misc.Ctxt, "Context switches should be > 0 since system is running")
	assert.Positive(t, misc.SysCalls, "System calls should be > 0 since system is running")
	assert.Positive(t, misc.Interrupts, "Interrupts should be > 0 since system is running")
}

func TestMiscAIX(t *testing.T) {
	misc, err := Misc()
	require.NoError(t, err)
	assert.NotNil(t, misc)
	assert.GreaterOrEqual(t, misc.ProcsTotal, 0)
	assert.GreaterOrEqual(t, misc.ProcsRunning, 0)
	assert.GreaterOrEqual(t, misc.ProcsBlocked, 0)
}
