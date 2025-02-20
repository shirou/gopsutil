// SPDX-License-Identifier: BSD-3-Clause
//go:build netbsd

package host

import (
	"context"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func HostIDWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func numProcs(_ context.Context) (uint64, error) {
	return 0, common.ErrNotImplementedError
}

func PlatformInformationWithContext(_ context.Context) (string, string, string, error) {
	platform := ""
	family := ""
	version := ""

	p, err := unix.Sysctl("kern.ostype")
	if err == nil {
		platform = strings.ToLower(p)
	}
	v, err := unix.Sysctl("kern.osrelease")
	if err == nil {
		version = strings.ToLower(v)
	}

	return platform, family, version, nil
}

func VirtualizationWithContext(_ context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}

func UsersWithContext(_ context.Context) ([]UserStat, error) {
	var ret []UserStat
	return ret, common.ErrNotImplementedError
}

func KernelVersionWithContext(ctx context.Context) (string, error) {
	_, _, version, err := PlatformInformationWithContext(ctx)
	return version, err
}
