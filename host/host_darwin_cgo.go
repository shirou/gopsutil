//go:build darwin && cgo
// +build darwin,cgo

package host

// #cgo LDFLAGS: -framework IOKit
// #include "smc_darwin.h"
import "C"
import "context"

func SensorsTemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
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
		C.CPU_EFFICIENCY_CORE_1,
		C.CPU_EFFICIENCY_CORE_2,
		C.CPU_PERFORMANCE_CORE_1,
		C.CPU_PERFORMANCE_CORE_2,
		C.CPU_PERFORMANCE_CORE_3,
		C.CPU_PERFORMANCE_CORE_4,
		C.CPU_PERFORMANCE_CORE_5,
		C.CPU_PERFORMANCE_CORE_6,
		C.CPU_PERFORMANCE_CORE_7,
		C.CPU_PERFORMANCE_CORE_8,
		C.GPU_1,
		C.GPU_2,
		C.GPU_3,
		C.GPU_4,
		C.AIRFLOW_LEFT,
		C.AIRFLOW_RIGHT,
		C.NAND,
		C.BATTERY_1,
		C.BATTERY_2,
		C.AIRPORT,
	}
	var temperatures []TemperatureStat

	C.gopsutil_v3_open_smc()
	defer C.gopsutil_v3_close_smc()

	for _, key := range temperatureKeys {
		temperatures = append(temperatures, TemperatureStat{
			SensorKey:   key,
			Temperature: float64(C.gopsutil_v3_get_temperature(C.CString(key))),
		})
	}
	return temperatures, nil
}
