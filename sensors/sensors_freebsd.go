// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package freebsd

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
