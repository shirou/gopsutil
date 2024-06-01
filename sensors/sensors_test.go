// SPDX-License-Identifier: BSD-3-Clause

package sensors

import (
	"fmt"
	"testing"
)

func TestTemperatureStat_String(t *testing.T) {
	v := TemperatureStat{
		SensorKey:   "CPU",
		Temperature: 1.1,
		High:        30.1,
		Critical:    0.1,
	}
	s := `{"sensorKey":"CPU","temperature":1.1,"sensorHigh":30.1,"sensorCritical":0.1}`
	if s != fmt.Sprintf("%v", v) {
		t.Errorf("TemperatureStat string is invalid, %v", fmt.Sprintf("%v", v))
	}
}
