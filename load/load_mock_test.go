// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package load

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cross-platform tests using mocked AIX command output
// These tests run on AIX systems, providing verification of parsing logic

func TestSystemCallsWithContextMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	ctx := context.Background()
	syscalls, err := SystemCallsWithContext(ctx)
	require.NoError(t, err)

	// Should extract 1083 from mock vmstat output
	assert.Equal(t, 1083, syscalls)
}

func TestInterruptsWithContextMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	ctx := context.Background()
	interrupts, err := InterruptsWithContext(ctx)
	require.NoError(t, err)

	// Should extract 9 from mock vmstat output
	assert.Equal(t, 9, interrupts)
}

func TestMiscWithContextMock(t *testing.T) {
	// Setup mock
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
	// Blocked: 1 (W) + 1 (Z) = 2
	assert.Equal(t, 10, misc.ProcsTotal)
	assert.Equal(t, 8, misc.ProcsRunning)
	assert.Equal(t, 2, misc.ProcsBlocked)

	// Should extract 669 from mock vmstat output
	assert.Equal(t, 669, misc.Ctxt)
}

func TestSystemCallsMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	syscalls, err := SystemCalls()
	require.NoError(t, err)
	assert.Equal(t, 1083, syscalls)
}

func TestInterruptsMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	interrupts, err := Interrupts()
	require.NoError(t, err)
	assert.Equal(t, 9, interrupts)
}

func TestMiscMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupSystemMetricsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	misc, err := Misc()
	require.NoError(t, err)
	assert.NotNil(t, misc)
	assert.Equal(t, 10, misc.ProcsTotal)
	assert.Equal(t, 8, misc.ProcsRunning)
	assert.Equal(t, 2, misc.ProcsBlocked)
	assert.Equal(t, 669, misc.Ctxt)
}
