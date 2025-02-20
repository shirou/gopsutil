// SPDX-License-Identifier: BSD-3-Clause
//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !solaris && !windows && !aix

package host

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func HostIDWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func numProcs(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

func BootTimeWithContext(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

func UptimeWithContext(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

func UsersWithContext(_ context.Context) ([]UserStat, error) {
	return []UserStat{}, common.ErrNotImplementedError
}

func VirtualizationWithContext(_ context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}

func KernelVersionWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func PlatformInformationWithContext(_ context.Context) (string, string, string, error) {
	return "", "", "", common.ErrNotImplementedError
}

func KernelArch() (string, error) {
	return "", common.ErrNotImplementedError
}
