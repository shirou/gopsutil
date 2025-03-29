// SPDX-License-Identifier: BSD-3-Clause
//go:build netbsd

package sensors

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TemperaturesWithContext(_ context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
