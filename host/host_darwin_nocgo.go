// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && !cgo

package host

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func SensorsTemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
