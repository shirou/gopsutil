// SPDX-License-Identifier: BSD-3-Clause
//go:build dragonfly

package host

import (
	"context"
	"os"
	"time"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// bootMarker is written once during boot, so its mtime is a boot timestamp.
const bootMarker = "/var/run/dmesg.boot"

func HostIDWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func numProcs(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

// BootTimeWithContext prefers kern.boottime. On KVM guests the kernel time base
// can be far off, leaving kern.boottime in the distant past (e.g. 1909); bootMarker
// is used in that case.
func BootTimeWithContext(_ context.Context) (uint64, error) {
	if boot, err := unix.SysctlTimeval("kern.boottime"); err == nil && boot.Sec > 0 {
		return uint64(boot.Sec), nil
	}
	info, err := os.Stat(bootMarker)
	if err != nil {
		return 0, err
	}
	return uint64(info.ModTime().Unix()), nil
}

func UptimeWithContext(ctx context.Context) (uint64, error) {
	boot, err := BootTimeWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(time.Now().Unix()) - boot, nil
}

func UsersWithContext(_ context.Context) ([]UserStat, error) {
	return []UserStat{}, common.ErrNotImplementedError
}

func VirtualizationWithContext(_ context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}

func KernelVersionWithContext(_ context.Context) (string, error) {
	return unix.Sysctl("kern.osrelease")
}

func PlatformInformationWithContext(ctx context.Context) (platform, family, version string, err error) {
	version, err = KernelVersionWithContext(ctx)
	return "dragonfly", "dragonfly", version, err
}

func KernelArch() (string, error) {
	var name unix.Utsname
	if err := unix.Uname(&name); err != nil {
		return "", err
	}
	return unix.ByteSliceToString(name.Machine[:]), nil
}
