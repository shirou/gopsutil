// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && cgo

package sensors

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Foundation -framework IOKit
// #include "smc_darwin.h"
// #include "darwin_arm_sensors.h"
import "C"
import (
	"bufio"
	"context"
	"math"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

func ReadTemperaturesArm() []TemperatureStat {
	cStr := C.getThermals()
	defer C.free(unsafe.Pointer(cStr))

	var stats []TemperatureStat
	goStr := C.GoString(cStr)
	scanner := bufio.NewScanner(strings.NewReader(goStr))
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), ":")
		if len(split) != 2 {
			continue
		}

		val, err := strconv.ParseFloat(split[1], 32)
		if err != nil {
			continue
		}

		sensorKey := strings.Split(split[0], " ")[0]

		val = math.Abs(val)

		stats = append(stats, TemperatureStat{
			SensorKey:   sensorKey,
			Temperature: float64(val),
		})
	}

	return stats
}

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	if runtime.GOARCH == "arm64" {
		return ReadTemperaturesArm(), nil
	}

	temperatureKeys := []string{
		C.AMBIENT_AIR_0,
		C.AMBIENT_AIR_1,
		C.CPU_0_DIODE,
		C.CPU_0_HEATSINK,
		C.CPU_0_PROXIMITY,
		C.ENCLOSURE_BASE_0,
		C.ENCLOSURE_BASE_1,
		C.ENCLOSURE_BASE_2,
		C.ENCLOSURE_BASE_3,
		C.GPU_0_DIODE,
		C.GPU_0_HEATSINK,
		C.GPU_0_PROXIMITY,
		C.HARD_DRIVE_BAY,
		C.MEMORY_SLOT_0,
		C.MEMORY_SLOTS_PROXIMITY,
		C.NORTHBRIDGE,
		C.NORTHBRIDGE_DIODE,
		C.NORTHBRIDGE_PROXIMITY,
		C.THUNDERBOLT_0,
		C.THUNDERBOLT_1,
		C.WIRELESS_MODULE,
	}
	var temperatures []TemperatureStat

	C.gopsutil_v4_open_smc()
	defer C.gopsutil_v4_close_smc()

	for _, key := range temperatureKeys {
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))
		temperatures = append(temperatures, TemperatureStat{
			SensorKey:   key,
			Temperature: float64(C.gopsutil_v4_get_temperature(ckey)),
		})
	}

	return temperatures, nil
}
