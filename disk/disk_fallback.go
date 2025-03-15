// SPDX-License-Identifier: BSD-3-Clause
//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !windows && !solaris && !aix

package disk

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func IOCountersWithContext(_ context.Context, _ ...string) (map[string]IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func PartitionsWithContext(_ context.Context, _ bool) ([]PartitionStat, error) {
	return []PartitionStat{}, common.ErrNotImplementedError
}

func UsageWithContext(_ context.Context, _ string) (*UsageStat, error) {
	return nil, common.ErrNotImplementedError
}

func SerialNumberWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}
