// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// from https://www.ibm.com/docs/en/aix/7.2?topic=files-utmph-file
const (
	user_PROCESS = 7 //nolint:revive //FIXME
)

func HostIDWithContext(ctx context.Context) (string, error) {
	out, err := invoke.CommandWithContext(ctx, "uname", "-u")
	if err != nil {
		return "", err
	}

	// The command always returns an extra newline, so we make use of Split() to get only the first line
	return strings.Split(string(out), "\n")[0], nil
}

func numProcs(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

func BootTimeWithContext(ctx context.Context) (btime uint64, err error) {
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
func UptimeWithContext(ctx context.Context) (uint64, error) {
	out, err := invoke.CommandWithContext(ctx, "uptime")
	if err != nil {
		return 0, err
	}

	return parseUptime(string(out)), nil
}

func parseUptime(uptime string) uint64 {
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

// This is a weak implementation due to the limitations on retrieving this data in AIX
func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	var ret []UserStat
	out, err := invoke.CommandWithContext(ctx, "w")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return []UserStat{}, common.ErrNotImplementedError
	}

	hf := strings.Fields(lines[1]) // headers
	for l := 2; l < len(lines); l++ {
		v := strings.Fields(lines[l]) // values
		us := &UserStat{}
		for i, header := range hf {
			// We're done in any of these use cases
			if i >= len(v) || v[0] == "-" {
				break
			}

			if t, err := strconv.ParseFloat(v[i], 64); err == nil {
				switch header {
				case `User`:
					us.User = strconv.FormatFloat(t, 'f', 1, 64)
				case `tty`:
					us.Terminal = strconv.FormatFloat(t, 'f', 1, 64)
				}
			}
		}

		// Valid User data, so append it
		ret = append(ret, *us)
	}

	return ret, nil
}

// Much of this function could be static. However, to be future proofed, I've made it call the OS for the information in all instances.
func PlatformInformationWithContext(ctx context.Context) (platform, family, version string, err error) {
	// Set the platform (which should always, and only be, "AIX") from `uname -s`
	out, err := invoke.CommandWithContext(ctx, "uname", "-s")
	if err != nil {
		return "", "", "", err
	}
	platform = strings.TrimRight(string(out), "\n")

	// Set the family
	family = strings.TrimRight(string(out), "\n")

	// Set the version
	out, err = invoke.CommandWithContext(ctx, "oslevel")
	if err != nil {
		return "", "", "", err
	}
	version = strings.TrimRight(string(out), "\n")

	return platform, family, version, nil
}

func KernelVersionWithContext(ctx context.Context) (version string, err error) {
	out, err := invoke.CommandWithContext(ctx, "oslevel", "-s")
	if err != nil {
		return "", err
	}
	version = strings.TrimRight(string(out), "\n")

	return version, nil
}

func KernelArch() (arch string, err error) {
	out, err := invoke.Command("bootinfo", "-y")
	if err != nil {
		return "", err
	}
	arch = strings.TrimRight(string(out), "\n")

	return arch, nil
}

func VirtualizationWithContext(_ context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}
