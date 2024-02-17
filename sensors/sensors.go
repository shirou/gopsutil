// SPDX-License-Identifier: BSD-3-Clause

package sensors

import (
	"context"
	"encoding/json"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type Warnings = common.Warnings

var invoke common.Invoker = common.Invoke{}

type TemperatureStat struct {
	SensorKey   string  `json:"sensorKey"`
	Temperature float64 `json:"temperature"`
	High        float64 `json:"sensorHigh"`
	Critical    float64 `json:"sensorCritical"`
}

func (t TemperatureStat) String() string {
	s, _ := json.Marshal(t)
	return string(s)
}

func SensorsTemperatures() ([]TemperatureStat, error) {
	return TemperaturesWithContext(context.Background())
}
