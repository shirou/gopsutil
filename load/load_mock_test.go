// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package load

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests using mocked AIX command output to verify parsing logic.

func TestMiscWithContextMock(t *testing.T) {
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	ctx := context.Background()
	misc, err := MiscWithContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, misc)

	// Mock data has: 7 A, 1 W, 1 I, 1 Z = 10 total
	// Running: 7 (A) + 1 (I) = 8
	// Blocked: 1 (W) = 1
	assert.Equal(t, 10, misc.ProcsTotal)
	assert.Equal(t, 8, misc.ProcsRunning)
	assert.Equal(t, 1, misc.ProcsBlocked)

	// Cumulative since-boot counters from vmstat -s mock output
	assert.Equal(t, 5842393706, misc.Ctxt)
	assert.Equal(t, 12918944607, misc.SysCalls)
	assert.Equal(t, 33412179, misc.Interrupts)
}

func TestMiscMock(t *testing.T) {
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	misc, err := Misc()
	require.NoError(t, err)
	assert.NotNil(t, misc)
	assert.Equal(t, 10, misc.ProcsTotal)
	assert.Equal(t, 8, misc.ProcsRunning)
	assert.Equal(t, 1, misc.ProcsBlocked)
	assert.Equal(t, 5842393706, misc.Ctxt)
	assert.Equal(t, 12918944607, misc.SysCalls)
	assert.Equal(t, 33412179, misc.Interrupts)
}
