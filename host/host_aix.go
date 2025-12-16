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

// Uses ps to get the elapsed time for PID 1 in DAYS-HOURS:MINUTES:SECONDS format.
// Examples of ps -o etimes -p 1 output:
// 124-01:40:39 (with days)
// 15:03:02 (without days, hours only)
// 01:02 (just-rebooted systems, minutes and seconds)
func UptimeWithContext(ctx context.Context) (uint64, error) {
	out, err := invoke.CommandWithContext(ctx, "ps", "-o", "etimes", "-p", "1")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, errors.New("ps output has fewer than 2 rows")
	}

	// Extract the etimes value from the second row, trimming whitespace
	etimes := strings.TrimSpace(lines[1])
	return parseUptime(etimes), nil
}

// Parses etimes output from ps command into total minutes.
// Handles formats like:
// - "124-01:40:39" (DAYS-HOURS:MINUTES:SECONDS)
// - "15:03:02" (HOURS:MINUTES:SECONDS)
// - "01:02" (MINUTES:SECONDS, from just-rebooted systems)
func parseUptime(etimes string) uint64 {
	var days, hours, mins, secs uint64

	// Check if days component is present (contains a dash)
	if strings.Contains(etimes, "-") {
		parts := strings.Split(etimes, "-")
		if len(parts) != 2 {
			return 0
		}

		var err error
		days, err = strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0
		}

		// Parse the HH:MM:SS portion
		etimes = parts[1]
	}

	// Parse time portions (either HH:MM:SS or MM:SS)
	timeParts := strings.Split(etimes, ":")
	switch len(timeParts) {
	case 3:
		// HH:MM:SS format
		var err error
		hours, err = strconv.ParseUint(timeParts[0], 10, 64)
		if err != nil {
			return 0
		}

		mins, err = strconv.ParseUint(timeParts[1], 10, 64)
		if err != nil {
			return 0
		}

		secs, err = strconv.ParseUint(timeParts[2], 10, 64)
		if err != nil {
			return 0
		}
	case 2:
		// MM:SS format (just-rebooted systems)
		var err error
		mins, err = strconv.ParseUint(timeParts[0], 10, 64)
		if err != nil {
			return 0
		}

		secs, err = strconv.ParseUint(timeParts[1], 10, 64)
		if err != nil {
			return 0
		}
	default:
		return 0
	}

	// Convert to total minutes
	totalMinutes := (days * 24 * 60) + (hours * 60) + mins + (secs / 60)
	return totalMinutes
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
