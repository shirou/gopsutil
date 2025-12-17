// SPDX-License-Identifier: BSD-3-Clause
//go:build !aix

package nfs

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// ClientStatsWithContext returns NFS client statistics
func ClientStatsWithContext(ctx context.Context) (*NFSClientStat, error) {
	return nil, common.ErrNotImplementedError
}

// ClientStats returns NFS client statistics
func ClientStats() (*NFSClientStat, error) {
	return nil, common.ErrNotImplementedError
}

// ServerStatsWithContext returns NFS server statistics
func ServerStatsWithContext(ctx context.Context) (*NFSServerStat, error) {
	return nil, common.ErrNotImplementedError
}

// ServerStats returns NFS server statistics
func ServerStats() (*NFSServerStat, error) {
	return nil, common.ErrNotImplementedError
}
