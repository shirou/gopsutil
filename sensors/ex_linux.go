// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package sensors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExTemperature represents Linux dependent temperature sensor data
type ExTemperature struct {
	SensorKey string  `json:"key"`
	Min       float64 `json:"min"`     // Temperature min value.
	Lowest    float64 `json:"lowest"`  // Historical minimum temperature
	Highest   float64 `json:"highest"` // Historical maximum temperature
}

type ExLinux struct{}

func NewExLinux() *ExLinux {
	return &ExLinux{}
}

func (ex *ExLinux) TemperatureWithContext(ctx context.Context) ([]ExTemperature, error) {
	var warns Warnings

	files, err := getTemperatureFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature files, %w", err)
	}

	temperatures := make([]ExTemperature, 0, len(files))
	for _, file := range files {
		var raw []byte

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

		// Add discovered temperature sensor to the list
		temperatures = append(temperatures, ExTemperature{
			SensorKey: name,
			Min:       optionalValueReadFromFile(basepath+"_min") / hostTemperatureScale,
			Lowest:    optionalValueReadFromFile(basepath+"_lowest") / hostTemperatureScale,
			Highest:   optionalValueReadFromFile(basepath+"_highest") / hostTemperatureScale,
		})
	}

	return temperatures, warns.Reference()
}
