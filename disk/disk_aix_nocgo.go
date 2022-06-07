//go:build aix && !cgo
// +build aix,!cgo

package disk

import (
	"context"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	return []PartitionStat{}, common.ErrNotImplementedError
}

func UsageWithContext(ctx context.Context, path string) (*UsageStat, error) {
	return nil, common.ErrNotImplementedError
}
