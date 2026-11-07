// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

// This is a weak implementation due to the limitations on retrieving this data in AIX
func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	var ret []UserStat
	out, err := getInvoker().CommandWithContext(ctx, "w")
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
		if len(v) == 0 || v[0] == "-" {
			continue
		}
		us := &UserStat{}
		for i, header := range hf {
			if i >= len(v) {
				break
			}
			switch header {
			case "User":
				us.User = v[i]
			case "tty":
				us.Terminal = v[i]
			}
		}

		// Valid User data, so append it
		ret = append(ret, *us)
	}

	return ret, nil
}

// bosRteLevelWithContext returns the VRMF level (version, release, modification
// and fix) of the bos.rte fileset along with its build date
func bosRteLevelWithContext(ctx context.Context) (version, release, mod, fix int, build string, err error) {
	// Use lslpp rather than oslevel which can fail for non-root users
	out, err := getInvoker().CommandWithContext(ctx, "lslpp", "-Lqc", "bos.rte")
	if err != nil {
		return 0, 0, 0, 0, "", err
	}

	// Colon-separated output without the header, e.g.:
	// bos:bos.rte:7.3.2.2: : :C:F:Base Operating System Runtime: : : : : : :0:0:/:2419
	// Package Name:Fileset:Level:State:PTF Id:Fix State:Type:Description:...:Install Path:Build Date
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	fields := strings.Split(line, ":")
	if len(fields) < 18 {
		return 0, 0, 0, 0, "", fmt.Errorf("unexpected lslpp output: %q", line)
	}

	vrmf := strings.Split(fields[2], ".")
	for len(vrmf) < 4 {
		vrmf = append(vrmf, "0")
	}

	nums := make([]int, 4)
	for i := range 4 {
		nums[i], err = strconv.Atoi(vrmf[i])
		if err != nil {
			return 0, 0, 0, 0, "", fmt.Errorf("unexpected bos.rte level %q: %w", fields[2], err)
		}
	}

	build = fields[len(fields)-1]
	if build == "" {
		return 0, 0, 0, 0, "", fmt.Errorf("no build date in lslpp output: %q", line)
	}

	return nums[0], nums[1], nums[2], nums[3], build, nil
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

	// Same as oslevel, e.g. "7.3.0.0".
	ver, rel, _, _, _, err := bosRteLevelWithContext(ctx)
	if err != nil {
		return "", "", "", err
	}
	version = fmt.Sprintf("%d.%d.0.0", ver, rel)

	return platform, family, version, nil
}

func KernelVersionWithContext(ctx context.Context) (version string, err error) {
	// Same as oslevel -s, e.g. "7300-02-02-2419": base level, technology
	// level, service pack and build date of bos.rte.
	ver, rel, mod, fix, build, err := bosRteLevelWithContext(ctx)
	if err != nil {
		return "", err
	}

	version = fmt.Sprintf("%d%d00-%02d-%02d-%s", ver, rel, mod, fix, build)

	return version, nil
}

func KernelArch() (arch string, err error) {
	out, err := getInvoker().Command("getconf", "KERNEL_BITMODE")
	if err != nil {
		out, err = getInvoker().Command("bootinfo", "-y")
		if err != nil {
			return "", err
		}
	}
	arch = strings.TrimRight(string(out), "\n")

	return arch, nil
}

func VirtualizationWithContext(_ context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}
