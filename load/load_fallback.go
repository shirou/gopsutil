// SPDX-License-Identifier: BSD-3-Clause
//go:build !darwin && !linux && !freebsd && !openbsd && !windows && !solaris && !aix

package load

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

func AvgWithContext(_ context.Context) (*AvgStat, error) {
	return nil, common.ErrNotImplementedError
}

func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

func MiscWithContext(_ context.Context) (*MiscStat, error) {
	return nil, common.ErrNotImplementedError
}
