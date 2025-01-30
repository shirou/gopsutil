// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package sensors

import (
	"context"
	"math"

	"github.com/yusufpapurcu/wmi"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type msAcpi_ThermalZoneTemperature struct {
	Active             bool
	CriticalTripPoint  uint32
	CurrentTemperature uint32
	InstanceName       string
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	var ret []TemperatureStat
	var dst []msAcpi_ThermalZoneTemperature
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithContext(ctx, q, &dst, nil, "root/wmi"); err != nil {
		return ret, err
	}

	for _, v := range dst {
		ts := TemperatureStat{
			SensorKey:   v.InstanceName,
			Temperature: kelvinToCelsius(v.CurrentTemperature, 2),
		}
		ret = append(ret, ts)
	}

	return ret, nil
}

func kelvinToCelsius(temp uint32, n int) float64 {
	// wmi return temperature Kelvin * 10, so need to divide the result by 10,
	// and then minus 273.15 to get °Celsius.
	t := float64(temp/10) - 273.15
	n10 := math.Pow10(n)
	return math.Trunc((t+0.5/n10)*n10) / n10
}
