// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package sensors

import (
	"context"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// deciKelvinToCelsius converts tenths of a Kelvin to degrees Celsius.
// The FreeBSD kernel stores -1 when a sensor value is missing or invalid.
func deciKelvinToCelsius(dK uint32) (float64, bool) {
	if int32(dK) < 0 {
		return 0, false
	}
	return float64(int32(dK)-2731) / 10.0, true
}

// sysctlDeciKelvin reads an "IK" sysctl and converts it to Celsius.
// The kernel stores -1 when the value is missing or invalid.
func sysctlDeciKelvin(name string) (float64, bool) {
	v, err := unix.SysctlUint32(name)
	if err != nil {
		return 0, false
	}
	return deciKelvinToCelsius(v)
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var ret []TemperatureStat

	// 1. CPU core temperatures via dev.cpu.%d.temperature
	if ncpu, err := unix.SysctlUint32("hw.ncpu"); err == nil {
		for i := 0; i < int(ncpu); i++ {
			temp, ok := sysctlDeciKelvin(fmt.Sprintf("dev.cpu.%d.temperature", i))
			if !ok {
				continue
			}

			var crit float64
			if tjmax, ok := sysctlDeciKelvin(fmt.Sprintf("dev.cpu.%d.coretemp.tjmax", i)); ok {
				crit = tjmax
			}

			ret = append(ret, TemperatureStat{
				SensorKey:   fmt.Sprintf("cpu%d", i),
				Temperature: temp,
				Critical:    crit,
			})
		}
	}

	// 2. ACPI thermal zones via hw.acpi.thermal.tz%d.temperature
	for i := 0; ; i++ {
		v, err := unix.SysctlUint32(fmt.Sprintf("hw.acpi.thermal.tz%d.temperature", i))
		if err != nil {
			break
		}
		temp, ok := deciKelvinToCelsius(v)
		if !ok {
			continue
		}

		var high float64
		if hot, ok := sysctlDeciKelvin(fmt.Sprintf("hw.acpi.thermal.tz%d._HOT", i)); ok {
			high = hot
		}

		var crit float64
		if crt, ok := sysctlDeciKelvin(fmt.Sprintf("hw.acpi.thermal.tz%d._CRT", i)); ok {
			crit = crt
		}

		ret = append(ret, TemperatureStat{
			SensorKey:   fmt.Sprintf("acpi_tz%d", i),
			Temperature: temp,
			High:        high,
			Critical:    crit,
		})
	}

	if len(ret) == 0 {
		return []TemperatureStat{}, common.ErrNotImplementedError
	}

	return ret, nil
}
