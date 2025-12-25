// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUptimeValidInput(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint64
	}{
		// Format: MINUTES:SECONDS (just-rebooted systems, hours dropped when 0)
		{"00:13", 13}, // 13 seconds
		{"01:00", 60}, // 1 minute
		{"01:02", 62}, // 1 minute, 2 seconds
		// Format: HOURS:MINUTES:SECONDS (no days, hours > 0)
		{"01:00:00", 3600},  // 1 hour
		{"05:00:00", 18000}, // 5 hours
		{"15:03:02", 54182}, // 15 hours, 3 minutes, 2 seconds
		// Format: DAYS-HOURS:MINUTES:SECONDS (with days)
		{"2-20:00:00", 244800},     // 2 days, 20 hours
		{"4-00:29:00", 347340},     // 4 days, 29 minutes
		{"83-18:29:00", 7237740},   // 83 days, 18 hours, 29 minutes
		{"124-01:40:39", 10719639}, // 124 days, 1 hour, 40 minutes, 39 seconds
	}
	for _, tc := range testCases {
		got := ParseUptime(tc.input)
		assert.Equalf(t, tc.expected, got, "ParseUptime(%q) = %v, want %v", tc.input, got, tc.expected)
	}
}

func TestParseUptimeInvalidInput(t *testing.T) {
	testCases := []string{
		"",             // blank
		"invalid",      // invalid string
		"1-2:3",        // incomplete time format after dash
		"abc-01:02:03", // non-numeric days
		"1-ab:02:03",   // non-numeric hours
	}

	for _, tc := range testCases {
		got := ParseUptime(tc)
		assert.Equalf(t, uint64(0), got, "ParseUptime(%q) expected zero to be returned, received %v", tc, got)
	}
}
