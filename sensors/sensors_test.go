// SPDX-License-Identifier: BSD-3-Clause

package sensors

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.Equalf(t, s, v.String(), "TemperatureStat string is invalid, %v", v.String())
}

func TestTemperatures(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skip CI")
	}
	v, err := SensorsTemperatures()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	assert.NotEmptyf(t, v, "Could not get temperature %v", v)
	t.Log(v)
}
