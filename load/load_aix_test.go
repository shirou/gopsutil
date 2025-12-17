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

	// Context switches should be positive (system has been running)
	assert.Greater(t, misc.Ctxt, 0, "Context switches should be > 0 since system is running")
}

func TestMiscAIX(t *testing.T) {
	// Test the non-context version
	misc, err := Misc()
	require.NoError(t, err)
	assert.NotNil(t, misc)

	// Process counts should be non-negative
	assert.GreaterOrEqual(t, misc.ProcsTotal, 0)
	assert.GreaterOrEqual(t, misc.ProcsRunning, 0)
	assert.GreaterOrEqual(t, misc.ProcsBlocked, 0)
}

func TestSystemCallsWithContext(t *testing.T) {
	ctx := context.Background()
	syscalls, err := SystemCallsWithContext(ctx)
	require.NoError(t, err)

	// System calls should be positive since system is running
	assert.Greater(t, syscalls, 0, "System calls should be > 0 since system is running")
}

func TestSystemCalls(t *testing.T) {
	syscalls, err := SystemCalls()
	require.NoError(t, err)

	// System calls should be positive
	assert.Greater(t, syscalls, 0)
}

func TestInterruptsWithContext(t *testing.T) {
	ctx := context.Background()
	interrupts, err := InterruptsWithContext(ctx)
	require.NoError(t, err)

	// Interrupts should be positive since system is running
	assert.Greater(t, interrupts, 0, "Interrupts should be > 0 since system is running")
}

func TestInterrupts(t *testing.T) {
	interrupts, err := Interrupts()
	require.NoError(t, err)

	// Interrupts should be positive
	assert.Greater(t, interrupts, 0)
}
