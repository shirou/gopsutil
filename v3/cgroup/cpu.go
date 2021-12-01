package cgroup

import (
	"context"
	"path"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/internal/common"
)

// CgroupCPU returns specified cgroup id CPU status.
func CgroupCPU() (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(context.Background(), getCgroupFilePath("", "cpuacct", "cpuacct.stat"))
}

// CgroupCPUUsage returns specified cgroup id CPU usage.
func CgroupCPUUsage() (float64, error) {
	return CgroupCPUUsageWithContext(context.Background(), getCgroupFilePath("", "cpuacct", "cpuacct.usage"))
}

func CgroupCPUWithContext(ctx context.Context, statfile string) (*CgroupCPUStat, error) {
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}

	ret := &CgroupCPUStat{}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if fields[0] == "user" {
			user, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.User = user / cpu.ClocksPerSec
			}
		}
		if fields[0] == "system" {
			system, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.System = system / cpu.ClocksPerSec
			}
		}
	}
	usage, err := CgroupCPUUsageWithContext(ctx, path.Join(path.Dir(statfile), "cpuacct.usage"))
	if err != nil {
		return nil, err
	}
	ret.Usage = usage
	return ret, nil
}

func CgroupCPUUsageWithContext(ctx context.Context, usagefile string) (float64, error) {
	lines, err := common.ReadLinesOffsetN(usagefile, 0, 1)
	if err != nil {
		return 0.0, err
	}

	ns, err := strconv.ParseFloat(lines[0], 64)
	if err != nil {
		return 0.0, err
	}

	return ns / nanoseconds, nil
}

