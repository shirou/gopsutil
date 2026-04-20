// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"strings"
	"sync"

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

// Static host information is cached because these values (hardware ID,
// platform name, OS version, kernel version, architecture) do not change
// at runtime. Caching avoids spawning subprocesses on repeated queries.
var (
	hostIDOnce sync.Once
	hostIDVal  string
	hostIDErr  error

	platformOnce sync.Once
	platformVal  string
	familyVal    string
	versionVal   string
	platformErr  error

	kernelVerOnce sync.Once
	kernelVerVal  string
	kernelVerErr  error

	kernelArchOnce sync.Once
	kernelArchVal  string
	kernelArchErr  error
)

func HostIDWithContext(ctx context.Context) (string, error) {
	hostIDOnce.Do(func() {
		out, err := getInvoker().CommandWithContext(ctx, "uname", "-u")
		if err != nil {
			hostIDErr = err
			return
		}
		hostIDVal = strings.Split(string(out), "\n")[0]
	})
	return hostIDVal, hostIDErr
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

// PlatformInformationWithContext returns the platform name, family, and OS
// version. These are immutable system identifiers cached after first query.
func PlatformInformationWithContext(ctx context.Context) (platform, family, version string, err error) {
	platformOnce.Do(func() {
		out, err := getInvoker().CommandWithContext(ctx, "uname", "-s")
		if err != nil {
			platformErr = err
			return
		}
		platformVal = strings.TrimRight(string(out), "\n")
		familyVal = platformVal

		out, err = getInvoker().CommandWithContext(ctx, "oslevel")
		if err != nil {
			platformErr = err
			return
		}
		versionVal = strings.TrimRight(string(out), "\n")
	})
	return platformVal, familyVal, versionVal, platformErr
}

// KernelVersionWithContext returns the kernel version (e.g., "7300-03-00-2446").
// This is an immutable system identifier cached after first query.
func KernelVersionWithContext(ctx context.Context) (version string, err error) {
	kernelVerOnce.Do(func() {
		out, err := getInvoker().CommandWithContext(ctx, "oslevel", "-s")
		if err != nil {
			kernelVerErr = err
			return
		}
		kernelVerVal = strings.TrimRight(string(out), "\n")
	})
	return kernelVerVal, kernelVerErr
}

// KernelArch returns the hardware architecture (e.g., "64").
// This is an immutable system identifier cached after first query.
func KernelArch() (arch string, err error) {
	kernelArchOnce.Do(func() {
		out, err := getInvoker().Command("bootinfo", "-y")
		if err != nil {
			kernelArchErr = err
			return
		}
		kernelArchVal = strings.TrimRight(string(out), "\n")
	})
	return kernelArchVal, kernelArchErr
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	// Check for WPAR (Workload Partition) first — most specific virtualization layer.
	// uname -W returns "0" if not in a WPAR, or the WPAR ID if inside one.
	out, err := getInvoker().CommandWithContext(ctx, "uname", "-W")
	if err == nil {
		wparID := strings.TrimSpace(string(out))
		if wparID != "" && wparID != "0" {
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
