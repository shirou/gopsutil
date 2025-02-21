// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package common

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

func BootTimeWithContext(ctx context.Context, invoke Invoker) (btime uint64, err error) {
	ut, err := UptimeWithContext(ctx)
	if err != nil {
		return 0, err
	}

	if ut <= 0 {
		return 0, errors.New("uptime was not set, so cannot calculate boot time from it")
	}

	ut *= 60
	return timeSince(ut), nil
}

// Parses result from uptime into minutes
// Some examples of uptime output that this command handles:
// 11:54AM   up 13 mins,  1 user,  load average: 2.78, 2.62, 1.79
// 12:41PM   up 1 hr,  1 user,  load average: 2.47, 2.85, 2.83
// 07:43PM   up 5 hrs,  1 user,  load average: 3.27, 2.91, 2.72
// 11:18:23  up 83 days, 18:29,  4 users,  load average: 0.16, 0.03, 0.01
// 08:47PM   up 2 days, 20 hrs, 1 user, load average: 2.47, 2.17, 2.17
// 01:16AM   up 4 days, 29 mins,  1 user,  load average: 2.29, 2.31, 2.21
func UptimeWithContext(ctx context.Context, invoke Invoker) (uint64, error) {
	out, err := invoke.CommandWithContext(ctx, "uptime")
	if err != nil {
		return 0, err
	}

	return ParseUptime(string(out[:])), nil
}

func ParseUptime(uptime string) uint64 {
	ut := strings.Fields(uptime)
	var days, hours, mins uint64
	var err error

	switch ut[3] {
	case "day,", "days,":
		days, err = strconv.ParseUint(ut[2], 10, 64)
		if err != nil {
			return 0
		}

		// day provided along with a single hour or hours
		// ie: up 2 days, 20 hrs,
		if ut[5] == "hr," || ut[5] == "hrs," {
			hours, err = strconv.ParseUint(ut[4], 10, 64)
			if err != nil {
				return 0
			}
		}

		// mins provided along with a single min or mins
		// ie: up 4 days, 29 mins,
		if ut[5] == "min," || ut[5] == "mins," {
			mins, err = strconv.ParseUint(ut[4], 10, 64)
			if err != nil {
				return 0
			}
		}

		// alternatively day provided with hh:mm
		// ie: up 83 days, 18:29
		if strings.Contains(ut[4], ":") {
			hm := strings.Split(ut[4], ":")
			hours, err = strconv.ParseUint(hm[0], 10, 64)
			if err != nil {
				return 0
			}
			mins, err = strconv.ParseUint(strings.Trim(hm[1], ","), 10, 64)
			if err != nil {
				return 0
			}
		}
	case "hr,", "hrs,":
		hours, err = strconv.ParseUint(ut[2], 10, 64)
		if err != nil {
			return 0
		}
	case "min,", "mins,":
		mins, err = strconv.ParseUint(ut[2], 10, 64)
		if err != nil {
			return 0
		}
	}

	return (days * 24 * 60) + (hours * 60) + mins
}
