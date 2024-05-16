//go:build aix
// +build aix

package host

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/internal/common"
)

// from https://www.ibm.com/docs/en/aix/7.2?topic=files-utmph-file
const (
	user_PROCESS = 7

	hostTemperatureScale = 1000.0 // Not part of the linked file, but kept just in case it becomes relevant
)

func HostIDWithContext(ctx context.Context) (string, error) {
	out, err := invoke.CommandWithContext(ctx, "uname", "-u")
	if err != nil {
		return "", err
	}

	// The command always returns an extra newline, so we make use of Split() to get only the first line
	return strings.Split(string(out[:]), "\n")[0]
}

func numProcs(ctx context.Context) (uint64, error) {
	return common.NumProcsWithContext(ctx)
}

func BootTimeWithContext(ctx context.Context) (btime uint64, err error) {
	ut, err := UptimeWithContext(ctx)
	if err != nil {
		return 0, err
	}

	if ut <= 0 {
		return 0, errors.New("Uptime was not set, so cannot calculate boot time from it.")
	}

	ut = ut * 60
	return timeSince(ut), nil
}

func UptimeWithContext(ctx context.Context) (uint64, error) {
	out, err := invoke.CommandWithContext(ctx, "uptime").Output()
	if err != nil {
		return 0, err
	}

	// Convert our uptime to a series of fields we can extract
	ut := strings.Fields(string(out[:]))

	// Convert the second field "Days" value to integer and roll it to minutes
	days, err := strconv.Atoi(ut[2])
	if err != nil {
		return 0, err
	}

	// Split field 4 into hours and minutes
	hm := strings.Split(ut[4], ":")
	hours, err := strconv.Atoi(hm[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(strings.Replace(hm[1], ",", "", -1))
	if err != nil {
		return 0, err
	}

	// Stack them all together as minutes
	total_time := (days * 24 * 60) + (hours * 60) + minutes

	return uint64(total_time), nil
}

// This is a weak implementation due to the limitations on retrieving this data in AIX
func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	var ret []UserStat
	out, err := invoke.CommandWithContext(ctx, "w").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return []UserStat{}, common.ErrNotImplementedError
	}

	hf := strings.Fields(lines[1]) // headers
	for l := 2; l < len(lines)-1; l++ {
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
					us.User = t
				case `tty`:
					us.Terminal = t
				}
			}
		}

		// Valid User data, so append it
		ret = append(ret, *us)
	}

	return ret, nil
}

// Much of this function could be static. However, to be future proofed, I've made it call the OS for the information in all instances.
func PlatformInformationWithContext(ctx context.Context) (platform string, family string, version string, err error) {
	// Set the platform (which should always, and only be, "AIX") from `uname -s`
	out, err := invoke.CommandWithContext(ctx, "uname", "-s").Output()
	if err != nil {
		return "", "", "", err
	}
	platform = string(out[:])

	// Set the family
	out, err = invoke.CommandWithContext(ctx, "bootinfo", "-p").Output()
	if err != nil {
		return "", "", "", err
	}
	// Family seems to always be the second field from this uname, so pull that out
	family = string(out[:])

	// Set the version
	out, err = invoke.CommandWithContext(ctx, "oslevel").Output()
	if err != nil {
		return "", "", "", err
	}
	version = string(out[:])

	return platform, family, version, nil
}

func KernelVersionWithContext(ctx context.Context) (version string, err error) {
	out, err := invoke.CommandWithContext(ctx, "oslevel", "-s").Output()
	if err != nil {
		return "", err
	}
	version = string(out[:])

	return version, nil
}

func KernelArch() (arch string, err error) {
	out, err := invoke.CommandWithContext(ctx, "bootinfo", "-y").Output()
	if err != nil {
		return "", err
	}
	arch = string(out[:])

	return arch, nil
}
