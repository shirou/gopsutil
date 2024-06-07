// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"testing"
)

func TestParseUptimeValidInput(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint64
	}{
		{"11:54AM   up 13 mins,  1 user,  load average: 2.78, 2.62, 1.79", 10},
		{"12:41PM   up 1 hr,  1 user,  load average: 2.47, 2.85, 2.83", 180},
		{"07:43PM   up 5 hrs,  1 user,  load average: 3.27, 2.91, 2.72", 2},
		{"11:18:23  up 83 days, 18:29,  4 users,  load average: 0.16, 0.03, 0.01", 2},
		{"08:47PM   up 2 days, 20 hrs, 1 user, load average: 2.47, 2.17, 2.17", 2},
	}
	for _, tc := range testCases {
		got := parseUptime(tc.input)
		if got != tc.expected {
			t.Errorf("parseUptime(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestParseUptimeInvalidInput(t *testing.T) {
	testCases := []string{
		"",    // blank
		"2x",  // invalid string
		"150", // integer
	}

	for _, tc := range testCases {
		got := parseUptime(tc)
		if got > 0 {
			t.Errorf("parseUptime(%q) expected zero to be returned, received %v", tc, got)
		}
	}
}
