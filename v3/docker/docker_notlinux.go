// +build !linux

package docker

import (
	"context"

	"github.com/shirou/gopsutil/v3/cgroup"
)

// GetDockerStat returns a list of Docker basic stats.
// This requires certain permission.
func GetDockerStat() ([]CgroupDockerStat, error) {
	return GetDockerStatWithContext(context.Background())
}

func GetDockerStatWithContext(ctx context.Context) ([]CgroupDockerStat, error) {
	return nil, ErrDockerNotAvailable
}

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	return GetDockerIDListWithContext(context.Background())
}

func GetDockerIDListWithContext(ctx context.Context) ([]string, error) {
	return nil, ErrDockerNotAvailable
}

func CgroupCPUDocker(containerID string) (*cgroup.CgroupCPUStat, error) {
	return CgroupCPUDockerWithContext(context.Background(), containerID)
}

func CgroupCPUDockerWithContext(ctx context.Context, containerID string) (*cgroup.CgroupCPUStat, error) {
	return  cgroup.CgroupCPUWithContext(ctx,
		getCgroupFilePath(containerID, "", "cpuacct", "cpuacct.stat"),
	)
}

func CgroupMemDocker(containerid string) (*cgroup.CgroupMemStat, error) {
	return CgroupMemDockerWithContext(context.Background(), containerid)
}

func CgroupMemDockerWithContext(ctx context.Context, containerID string) (*cgroup.CgroupMemStat, error) {
	return cgroup.CgroupMemWithContext(ctx,
		getCgroupFilePath(containerID, "", "memory", "memory.stat"),
	)
}
