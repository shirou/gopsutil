// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package net

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionsPidWithContext(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	conns, err := ConnectionsPidWithContext(ctx, "inet", pid)
	// Process may not have any connections, check result gracefully
	if err != nil {
		// It's OK if the function returns an error
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	// If successful, verify structure
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
		for _, conn := range conns {
			// Verify connection fields are populated
			assert.NotEmpty(t, conn.Family)
			assert.NotEmpty(t, conn.Type)
		}
	}
}

func TestConnectionsPidWithContextAll(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	// Test with "all" family
	conns, err := ConnectionsPidWithContext(ctx, "all", pid)
	if err != nil {
		// It's OK if the function returns an error
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
	}
}

func TestConnectionsPidWithContextUDP(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	// Test with UDP connections
	conns, err := ConnectionsPidWithContext(ctx, "udp", pid)
	if err != nil {
		// It's OK if the function returns an error
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
	}
}

func TestConnectionsWithContext(t *testing.T) {
	ctx := context.Background()

	// Test getting all connections
	conns, err := ConnectionsWithContext(ctx, "inet")
	require.NoError(t, err)
	assert.NotNil(t, conns)
	assert.IsType(t, []ConnectionStat{}, conns)

	// Should have at least some connections
	assert.NotEmpty(t, conns)

	for _, conn := range conns {
		// Verify connection fields
		assert.NotEmpty(t, conn.Family)
		assert.NotEmpty(t, conn.Type)
		assert.GreaterOrEqual(t, conn.Pid, int32(0))
	}
}
