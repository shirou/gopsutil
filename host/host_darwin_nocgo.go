// +build darwin
// +build !cgo

package host

import (
	"context"

	"github.com/redhatxl/gopsutil/internal/common"
)

func SensorsTemperatures() ([]TemperatureStat, error) {
	return SensorsTemperaturesWithContext(context.Background())
}

func SensorsTemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	return []TemperatureStat{}, common.ErrNotImplementedError
}
