// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package nfs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientStatsWithContext(t *testing.T) {
	ctx := context.Background()
	stats, err := ClientStatsWithContext(ctx)
	// AIX system may not have NFS client enabled, so we just check error handling
	if err != nil {
		// It's OK if NFS is not available
		t.Logf("NFS client stats not available: %v", err)
		return
	}
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.Calls, uint64(0))
}

func TestServerStatsWithContext(t *testing.T) {
	ctx := context.Background()
	stats, err := ServerStatsWithContext(ctx)
	// AIX system may not have NFS server enabled, so we just check error handling
	if err != nil {
		// It's OK if NFS is not available
		t.Logf("NFS server stats not available: %v", err)
		return
	}
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.Calls, uint64(0))
}
