// SPDX-License-Identifier: BSD-3-Clause
//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !solaris && !windows && !aix

package sensors

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
