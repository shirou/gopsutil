// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
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

// aixUtmp matches the AIX /etc/utmp binary record layout (see /usr/include/utmp.h).
// Reading utmp directly is ~180x faster than spawning `who` and avoids locale
// dependencies when parsing timestamps — ut_time is an epoch value.
type aixUtmp struct {
	User     [256]byte // ut_user: login name
	ID       [14]byte  // ut_id: inittab id
	Line     [64]byte  // ut_line: device name (pts/0, etc.)
	Pid      int32     // ut_pid
	Type     int16     // ut_type (7 = USER_PROCESS)
	Time     int64     // ut_time: epoch seconds (time64_t)
	Exit     [4]byte   // ut_exit: termination/exit status
	Host     [256]byte // ut_host: remote host
	Pad      [4]byte   // __dbl_word_pad
	Reserved [32]byte  // __reservedA[2] + __reservedV[6]
}

// UsersWithContext returns currently logged-in users by reading /etc/utmp directly.
// This avoids spawning a subprocess and eliminates locale dependencies for
// timestamp parsing — the utmp struct contains epoch seconds in ut_time.
func UsersWithContext(_ context.Context) ([]UserStat, error) {
	f, err := os.Open("/etc/utmp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ret []UserStat
	for {
		var entry aixUtmp
		err := binary.Read(f, binary.BigEndian, &entry)
		if err != nil {
			break // EOF or read error
		}

		// Only include active user sessions (ut_type == USER_PROCESS)
		if entry.Type != user_PROCESS {
			continue
		}

		user := strings.TrimRight(string(bytes.TrimRight(entry.User[:], "\x00")), " ")
		if user == "" {
			continue
		}

		us := UserStat{
			User:     user,
			Terminal: string(bytes.TrimRight(entry.Line[:], "\x00")),
			Host:     string(bytes.TrimRight(entry.Host[:], "\x00")),
			Started:  int(entry.Time),
		}
		ret = append(ret, us)
	}

	return ret, nil
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
