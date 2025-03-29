// SPDX-License-Identifier: BSD-3-Clause
//go:build !linux

package docker

import (
	"context"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// GetDockerStat returns a list of Docker basic stats.
// This requires certain permission.
func GetDockerStat() ([]CgroupDockerStat, error) {
	return GetDockerStatWithContext(context.Background())
}

func GetDockerStatWithContext(_ context.Context) ([]CgroupDockerStat, error) {
	return nil, ErrDockerNotAvailable
}

// GetDockerIDList returns a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	return GetDockerIDListWithContext(context.Background())
}

func GetDockerIDListWithContext(_ context.Context) ([]string, error) {
	return nil, ErrDockerNotAvailable
}

// CgroupCPU returns specified cgroup id CPU status.
// containerid is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerid = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerid string, base string) (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(context.Background(), containerid, base)
}

func CgroupCPUWithContext(_ context.Context, _ string, _ string) (*CgroupCPUStat, error) {
	return nil, ErrCgroupNotAvailable
}

func CgroupCPUDocker(containerid string) (*CgroupCPUStat, error) {
	return CgroupCPUDockerWithContext(context.Background(), containerid)
}

func CgroupCPUDockerWithContext(ctx context.Context, containerid string) (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(ctx, containerid, common.HostSysWithContext(ctx, "fs/cgroup/cpuacct/docker"))
}

func CgroupMem(containerid string, base string) (*CgroupMemStat, error) {
	return CgroupMemWithContext(context.Background(), containerid, base)
}

func CgroupMemWithContext(_ context.Context, _ string, _ string) (*CgroupMemStat, error) {
	return nil, ErrCgroupNotAvailable
}

func CgroupMemDocker(containerid string) (*CgroupMemStat, error) {
	return CgroupMemDockerWithContext(context.Background(), containerid)
}

func CgroupMemDockerWithContext(ctx context.Context, containerid string) (*CgroupMemStat, error) {
	return CgroupMemWithContext(ctx, containerid, common.HostSysWithContext(ctx, "fs/cgroup/memory/docker"))
}
