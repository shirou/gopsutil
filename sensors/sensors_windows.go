// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package sensors

import (
	"context"
	"math"
	"strings"

	"github.com/yusufpapurcu/wmi"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type msAcpi_ThermalZoneTemperature struct { //nolint:revive //FIXME
	Active             bool
	CriticalTripPoint  uint32
	CurrentTemperature uint32
	InstanceName       string
}

type win32_PerfFormattedData_Counters_ThermalZoneInformation struct { //nolint:revive //FIXME
	Name                     string
	HighPrecisionTemperature uint32
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	ret, err := acpiQuery(ctx)
	if strings.Contains(err.Error(), "Access denied") {
		ret, err = win32Query(ctx)
	}

	return ret, err
}

func acpiQuery(ctx context.Context) ([]TemperatureStat, error) {
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

func win32Query(ctx context.Context) ([]TemperatureStat, error) {
	var ret []TemperatureStat
	var dst []win32_PerfFormattedData_Counters_ThermalZoneInformation
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithContext(ctx, q, &dst, nil); err != nil {
		return ret, err
	}

	for _, v := range dst {
		ts := TemperatureStat{
			SensorKey:   v.Name,
			Temperature: kelvinToCelsius(v.HighPrecisionTemperature, 2),
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
