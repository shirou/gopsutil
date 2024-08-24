// SPDX-License-Identifier: BSD-3-Clause
//go:build openbsd

package sensors

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
