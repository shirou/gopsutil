// +build linux

package docker

import (
	"context"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v3/cgroup"
)

// GetDockerStat returns a list of Docker basic stats.
// This requires certain permission.
func GetDockerStat() ([]CgroupDockerStat, error) {
	return GetDockerStatWithContext(context.Background())
}

func GetDockerStatWithContext(ctx context.Context) ([]CgroupDockerStat, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, ErrDockerNotAvailable
	}

	out, err := invoke.CommandWithContext(ctx, path, "ps", "-a", "--no-trunc", "--format", "{{.ID}}|{{.Image}}|{{.Names}}|{{.Status}}")
	if err != nil {
		return []CgroupDockerStat{}, err
	}
	lines := strings.Split(string(out), "\n")
	ret := make([]CgroupDockerStat, 0, len(lines))

	for _, l := range lines {
		if l == "" {
			continue
		}
		cols := strings.Split(l, "|")
		if len(cols) != 4 {
			continue
		}
		names := strings.Split(cols[2], ",")
		stat := CgroupDockerStat{
			ContainerID: cols[0],
			Name:        names[0],
			Image:       cols[1],
			Status:      cols[3],
			Running:     strings.Contains(cols[3], "Up"),
		}
		ret = append(ret, stat)
	}

	return ret, nil
}

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	return GetDockerIDListWithContext(context.Background())
}

func GetDockerIDListWithContext(ctx context.Context) ([]string, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, ErrDockerNotAvailable
	}

	out, err := invoke.CommandWithContext(ctx, path, "ps", "-q", "--no-trunc")
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(out), "\n")
	ret := make([]string, 0, len(lines))

	for _, l := range lines {
		if l == "" {
			continue
		}
		ret = append(ret, l)
	}

	return ret, nil
}

func CgroupCPUDocker(containerID string) (*cgroup.CgroupCPUStat, error) {
	return CgroupCPUDockerWithContext(context.Background(), containerID)
}

func CgroupCPUUsageDocker(containerID string) (float64, error) {
	return CgroupCPUDockerUsageWithContext(context.Background(), containerID)
}

func CgroupCPUDockerWithContext(ctx context.Context, containerID string) (*cgroup.CgroupCPUStat, error) {
	return cgroup.CgroupCPUWithContext(ctx,
		getCgroupFilePath(containerID, "", "cpuacct", "cpuacct.stat"),
	)
}

func CgroupCPUDockerUsageWithContext(ctx context.Context, containerID string) (float64, error) {
	return cgroup.CgroupCPUUsageWithContext(ctx,
		getCgroupFilePath(containerID, "", "cpuacct", "cpuacct.usage"),
	)
}

func CgroupMemDocker(containerID string) (*cgroup.CgroupMemStat, error) {
	return CgroupMemDockerWithContext(context.Background(), containerID)
}

func CgroupMemDockerWithContext(ctx context.Context, containerID string) (*cgroup.CgroupMemStat, error) {
	return cgroup.CgroupMemWithContext(ctx,
		getCgroupFilePath(containerID, "", "memory", "memory.stat"),
	)
}
