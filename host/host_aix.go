// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// from https://www.ibm.com/docs/en/aix/7.2?topic=files-utmph-file
const (
	user_PROCESS = 7 //nolint:revive //FIXME
)

// testInvoker is used for dependency injection in tests
var testInvoker common.Invoker

// getInvoker returns the test invoker if set, otherwise returns the default
func getInvoker() common.Invoker {
	if testInvoker != nil {
		return testInvoker
	}
	return invoke
}

func HostIDWithContext(ctx context.Context) (string, error) {
	out, err := getInvoker().CommandWithContext(ctx, "uname", "-u")
	if err != nil {
		return "", err
	}

	// The command always returns an extra newline, so we make use of Split() to get only the first line
	return strings.Split(string(out), "\n")[0], nil
}

func BootTimeWithContext(ctx context.Context) (btime uint64, err error) {
	return common.BootTimeWithContext(ctx, getInvoker())
}

// Uses ps to get the elapsed time for PID 1 in DAYS-HOURS:MINUTES:SECONDS format.
func UptimeWithContext(ctx context.Context) (uint64, error) {
	return common.UptimeWithContext(ctx, getInvoker())
}

// UsersWithContext returns a list of currently logged-in users by parsing `who` output.
// Output format: root        pts/0       Feb 27 06:58     (24.236.207.124)
func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	out, err := getInvoker().CommandWithContext(ctx, "who")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []UserStat{}, nil
	}

	now := time.Now()
	var ret []UserStat
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		us := UserStat{
			User:     fields[0],
			Terminal: fields[1],
			Started:  parseWhoTimestamp(fields[2], fields[3], fields[4], now),
		}
		if len(fields) >= 6 {
			us.Host = strings.Trim(fields[5], "()")
		}

		ret = append(ret, us)
	}

	return ret, nil
}

// parseWhoTimestamp converts the month, day, and time fields from `who` output
// (e.g. "Feb", "27", "06:58") into a unix timestamp. The year is inferred from
// the current time, with a correction for the year boundary: if the login month
// is December but the current month is January, the login happened last year.
func parseWhoTimestamp(month, day, hhmm string, now time.Time) int {
	loginTime, err := time.Parse("Jan 2 15:04", month+" "+day+" "+hhmm)
	if err != nil {
		return 0
	}

	year := now.Year()
	if loginTime.Month() == time.December && now.Month() == time.January {
		year--
	}

	loginTime = time.Date(year, loginTime.Month(), loginTime.Day(),
		loginTime.Hour(), loginTime.Minute(), 0, 0, time.Local)
	return int(loginTime.Unix())
}

// Much of this function could be static. However, to be future proofed, I've made it call the OS for the information in all instances.
func PlatformInformationWithContext(ctx context.Context) (platform, family, version string, err error) {
	// Set the platform (which should always, and only be, "AIX") from `uname -s`
	out, err := getInvoker().CommandWithContext(ctx, "uname", "-s")
	if err != nil {
		return "", "", "", err
	}
	platform = strings.TrimRight(string(out), "\n")

	// Set the family
	family = strings.TrimRight(string(out), "\n")

	// Set the version
	out, err = getInvoker().CommandWithContext(ctx, "oslevel")
	if err != nil {
		return "", "", "", err
	}
	version = strings.TrimRight(string(out), "\n")

	return platform, family, version, nil
}

func KernelVersionWithContext(ctx context.Context) (version string, err error) {
	out, err := getInvoker().CommandWithContext(ctx, "oslevel", "-s")
	if err != nil {
		return "", err
	}
	version = strings.TrimRight(string(out), "\n")

	return version, nil
}

func KernelArch() (arch string, err error) {
	out, err := getInvoker().Command("bootinfo", "-y")
	if err != nil {
		return "", err
	}
	arch = strings.TrimRight(string(out), "\n")

	return arch, nil
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	// Check for WPAR (Workload Partition) first — most specific virtualization layer.
	// uname -W returns "0" if not in a WPAR, or the WPAR ID if inside one.
	out, err := getInvoker().CommandWithContext(ctx, "uname", "-W")
	if err == nil {
		wparID := strings.TrimSpace(string(out))
		if wparID != "0" {
			return "wpar", "guest", nil
		}
	}

	// Check for LPAR (Logical Partition) via PowerVM.
	// uname -L returns "<id> <name>", e.g. "25 soaix422". If name is "NULL", no LPAR.
	out, err = getInvoker().CommandWithContext(ctx, "uname", "-L")
	if err == nil {
		fields := strings.Fields(strings.TrimSpace(string(out)))
		if len(fields) >= 2 && fields[1] != "NULL" {
			return "powervm", "guest", nil
		}
	}

	return "", "", nil
}
