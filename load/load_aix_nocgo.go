//go:build aix && !cgo
// +build aix,!cgo

package load

import (
	"context"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	return nil, common.ErrNotImplementedError
}

func MiscWithContext(ctx context.Context) (*MiscStat, error) {
	return nil, common.ErrNotImplementedError
}
