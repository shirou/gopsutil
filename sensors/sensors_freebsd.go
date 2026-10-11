// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package sensors

import (
	"context"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// DeciKelvinToCelsius converts tenths of a Kelvin to degrees Celsius.
func DeciKelvinToCelsius(dK uint32) float64 {
	return float64(int64(dK)-2731) / 10.0
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var ret []TemperatureStat

	// 1. CPU core temperatures via dev.cpu.%d.temperature
	ncpu, err := unix.SysctlUint32("hw.ncpu")
	cpuLimit := 1024
	if err == nil && ncpu > 0 {
		cpuLimit = int(ncpu)
	}

	for i := 0; i < cpuLimit; i++ {
		temp, err := unix.SysctlUint32(fmt.Sprintf("dev.cpu.%d.temperature", i))
		if err != nil {
			if ncpu == 0 {
				break
			}
			continue
		}

		var crit float64
		if tjmax, err := unix.SysctlUint32(fmt.Sprintf("dev.cpu.%d.coretemp.tjmax", i)); err == nil {
			crit = DeciKelvinToCelsius(tjmax)
		}

		ret = append(ret, TemperatureStat{
			SensorKey:   fmt.Sprintf("cpu%d", i),
			Temperature: DeciKelvinToCelsius(temp),
			Critical:    crit,
		})
	}

	// 2. ACPI thermal zones via hw.acpi.thermal.tz%d.temperature
	for i := 0; ; i++ {
		temp, err := unix.SysctlUint32(fmt.Sprintf("hw.acpi.thermal.tz%d.temperature", i))
		if err != nil {
			break
		}

		var crit float64
		if crt, err := unix.SysctlUint32(fmt.Sprintf("hw.acpi.thermal.tz%d._CRT", i)); err == nil {
			crit = DeciKelvinToCelsius(crt)
		}

		ret = append(ret, TemperatureStat{
			SensorKey:   fmt.Sprintf("acpi_tz%d", i),
			Temperature: DeciKelvinToCelsius(temp),
			Critical:    crit,
		})
	}

	if len(ret) == 0 {
		return []TemperatureStat{}, common.ErrNotImplementedError
	}

	return ret, nil
}
