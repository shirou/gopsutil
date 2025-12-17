// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests using mocked AIX command output
// These tests run on AIX systems, providing verification of parsing logic

func TestFDLimitsWithContextMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupFDLimitsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	ctx := context.Background()
	soft, hard, err := FDLimitsWithContext(ctx)
	require.NoError(t, err)

	// Should extract 2048 from mock ulimit -n
	assert.Equal(t, uint64(2048), soft)

	// Should recognize "unlimited" and return max int64
	assert.Equal(t, uint64(1<<63-1), hard)

	// Hard should be >= soft
	assert.GreaterOrEqual(t, hard, soft)
}

func TestFDLimitsMock(t *testing.T) {
	// Setup mock
	mock := NewMockInvoker()
	mock.SetupFDLimitsMock()
	testInvoker = mock
	defer func() { testInvoker = nil }()

	soft, hard, err := FDLimits()
	require.NoError(t, err)

	assert.Equal(t, uint64(2048), soft)
	assert.Equal(t, uint64(1<<63-1), hard)
}
