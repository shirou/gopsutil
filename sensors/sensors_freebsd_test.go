// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd

package sensors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestDeciKelvinToCelsius(t *testing.T) {
	testCases := []struct {
		name     string
		dK       uint32
		expected float64
	}{
		{
			name:     "negative celsius (-10.0 C)",
			dK:       2631,
			expected: -10.0,
		},
		{
			name:     "freezing point 0.0 C",
			dK:       2731,
			expected: 0.0,
		},
		{
			name:     "acpi thermal zone 27.9 C",
			dK:       3010,
			expected: 27.9,
		},
		{
			name:     "typical core temp 47.0 C",
			dK:       3201,
			expected: 47.0,
		},
		{
			name:     "intel coretemp tjmax 105.0 C",
			dK:       3781,
			expected: 105.0,
		},
		{
			name:     "acpi critical trip point 110.1 C",
			dK:       3832,
			expected: 110.1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := DeciKelvinToCelsius(tc.dK)
			assert.InDelta(t, tc.expected, actual, 0.0001)
		})
	}
}

func TestTemperaturesWithContext_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := TemperaturesWithContext(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestTemperaturesWithContext_Live(t *testing.T) {
	stats, err := TemperaturesWithContext(context.Background())
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("No FreeBSD coretemp or acpi_thermal sensors available on host")
	}
	require.NoError(t, err)
	assert.NotEmpty(t, stats)

	for _, stat := range stats {
		assert.NotEmpty(t, stat.SensorKey)
		t.Logf("SensorKey: %s, Temperature: %.1f C, Critical: %.1f C", stat.SensorKey, stat.Temperature, stat.Critical)
	}
}
