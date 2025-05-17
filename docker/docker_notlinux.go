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
// containerID is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerID = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerID, base string) (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(context.Background(), containerID, base)
}

func CgroupCPUWithContext(_ context.Context, _, _ string) (*CgroupCPUStat, error) {
	return nil, ErrCgroupNotAvailable
}

func CgroupCPUDocker(containerID string) (*CgroupCPUStat, error) {
	return CgroupCPUDockerWithContext(context.Background(), containerID)
}

func CgroupCPUDockerWithContext(ctx context.Context, containerID string) (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(ctx, containerID, common.HostSysWithContext(ctx, "fs/cgroup/cpuacct/docker"))
}

func CgroupMem(containerID, base string) (*CgroupMemStat, error) {
	return CgroupMemWithContext(context.Background(), containerID, base)
}

func CgroupMemWithContext(_ context.Context, _, _ string) (*CgroupMemStat, error) {
	return nil, ErrCgroupNotAvailable
}

func CgroupMemDocker(containerID string) (*CgroupMemStat, error) {
	return CgroupMemDockerWithContext(context.Background(), containerID)
}

func CgroupMemDockerWithContext(ctx context.Context, containerID string) (*CgroupMemStat, error) {
	return CgroupMemWithContext(ctx, containerID, common.HostSysWithContext(ctx, "fs/cgroup/memory/docker"))
}
