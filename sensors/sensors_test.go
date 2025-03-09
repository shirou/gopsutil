// SPDX-License-Identifier: BSD-3-Clause

package sensors

import (
	"fmt"
	"os"
	"testing"

	"github.com/shirou/gopsutil/v4/internal/common"
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

func TestTemperatures(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skip CI")
	}
	v, err := SensorsTemperatures()
	common.SkipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("Could not get temperature %v", v)
	}
	t.Log(v)
}
