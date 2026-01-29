// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package sensors

import (
	"context"

	"github.com/yusufpapurcu/wmi"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type msAcpi_ThermalZoneTemperature struct { //nolint:revive //FIXME
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
			Temperature: kelvinToCelsius(v.CurrentTemperature),
		}
		ret = append(ret, ts)
	}

	return ret, nil
}

func kelvinToCelsius(temp uint32) float64 {
	// wmi return temperature Kelvin * 10, so need to divide the result by 10,
	// and then minus 273.15 to get °Celsius.
	return (float64(temp*10) - 27315) / 100
}
