// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package sensors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// from utmp.h
const (
	hostTemperatureScale = 1000.0
)

func TemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	var warns Warnings

	files, err := getTemperatureFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature files, %w", err)
	}

	if len(files) == 0 { // handle distributions without hwmon, like raspbian #391, parse legacy thermal_zone files
		files, err = filepath.Glob(common.HostSysWithContext(ctx, "/class/thermal/thermal_zone*/"))
		if err != nil {
			return nil, err
		}
		temperatures := make([]TemperatureStat, 0, len(files))

		for _, file := range files {
			// Get the name of the temperature you are reading
			name, err := os.ReadFile(filepath.Join(file, "type"))
			if err != nil {
				warns.Add(err)
				continue
			}
			// Get the temperature reading
			current, err := os.ReadFile(filepath.Join(file, "temp"))
			if err != nil {
				warns.Add(err)
				continue
			}
			temperature, err := strconv.ParseInt(strings.TrimSpace(string(current)), 10, 64)
			if err != nil {
				warns.Add(err)
				continue
			}

			temperatures = append(temperatures, TemperatureStat{
				SensorKey:   strings.TrimSpace(string(name)),
				Temperature: float64(temperature) / 1000.0,
			})
		}
		return temperatures, warns.Reference()
	}

	temperatures := make([]TemperatureStat, 0, len(files))

	// example directory
	// device/           temp1_crit_alarm  temp2_crit_alarm  temp3_crit_alarm  temp4_crit_alarm  temp5_crit_alarm  temp6_crit_alarm  temp7_crit_alarm
	// name              temp1_input       temp2_input       temp3_input       temp4_input       temp5_input       temp6_input       temp7_input
	// power/            temp1_label       temp2_label       temp3_label       temp4_label       temp5_label       temp6_label       temp7_label
	// subsystem/        temp1_max         temp2_max         temp3_max         temp4_max         temp5_max         temp6_max         temp7_max
	// temp1_crit        temp2_crit        temp3_crit        temp4_crit        temp5_crit        temp6_crit        temp7_crit        uevent
	for _, file := range files {
		var raw []byte

		var temperature float64

		// Get the base directory location
		directory := filepath.Dir(file)

		// Get the base filename prefix like temp1
		basename := strings.Split(filepath.Base(file), "_")[0]

		// Get the base path like <dir>/temp1
		basepath := filepath.Join(directory, basename)

		// Get the label of the temperature you are reading
		label := ""

		if raw, _ = os.ReadFile(basepath + "_label"); len(raw) != 0 {
			// Format the label from "Core 0" to "core_0"
			label = strings.Join(strings.Split(strings.TrimSpace(strings.ToLower(string(raw))), " "), "_")
		}

		// Get the name of the temperature you are reading
		if raw, err = os.ReadFile(filepath.Join(directory, "name")); err != nil {
			warns.Add(err)
			continue
		}

		name := strings.TrimSpace(string(raw))

		if label != "" {
			name = name + "_" + label
		}

		// Get the temperature reading
		if raw, err = os.ReadFile(file); err != nil {
			warns.Add(err)
			continue
		}

		if temperature, err = strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); err != nil {
			warns.Add(err)
			continue
		}

		// Add discovered temperature sensor to the list
		temperatures = append(temperatures, TemperatureStat{
			SensorKey:   name,
			Temperature: temperature / hostTemperatureScale,
			High:        optionalValueReadFromFile(basepath+"_max") / hostTemperatureScale,
			Critical:    optionalValueReadFromFile(basepath+"_crit") / hostTemperatureScale,
		})
	}

	return temperatures, warns.Reference()
}

func getTemperatureFiles(ctx context.Context) ([]string, error) {
	var files []string
	var err error

	// Only the temp*_input file provides current temperature
	// value in millidegree Celsius as reported by the temperature to the device:
	// https://www.kernel.org/doc/Documentation/hwmon/sysfs-interface
	if files, err = filepath.Glob(common.HostSysWithContext(ctx, "/class/hwmon/hwmon*/temp*_input")); err != nil {
		return nil, err
	}

	if len(files) == 0 {
		// CentOS has an intermediate /device directory:
		// https://github.com/giampaolo/psutil/issues/971
		if files, err = filepath.Glob(common.HostSysWithContext(ctx, "/class/hwmon/hwmon*/device/temp*_input")); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func optionalValueReadFromFile(filename string) float64 {
	var raw []byte

	var err error

	var value float64

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return 0
	}

	if raw, err = os.ReadFile(filename); err != nil {
		return 0
	}

	if value, err = strconv.ParseFloat(strings.TrimSpace(string(raw)), 64); err != nil {
		return 0
	}

	return value
}
