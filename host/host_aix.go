// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
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
	return common.BootTimeWithContext(ctx, invoke)
}

// Uses ps to get the elapsed time for PID 1 in DAYS-HOURS:MINUTES:SECONDS format.
func UptimeWithContext(ctx context.Context) (uint64, error) {
	return common.UptimeWithContext(ctx, invoke)
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
