// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package docker

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/internal/common"
)

// microseconds is the unit of the counters in the cgroup v2 cpu.stat file.
const microseconds = 1e6

// GetDockerStat returns a list of Docker basic stats.
// This requires certain permission.
func GetDockerStat() ([]CgroupDockerStat, error) {
	return GetDockerStatWithContext(context.Background())
}

func GetDockerStatWithContext(ctx context.Context) ([]CgroupDockerStat, error) {
	out, err := invoke.CommandWithContext(ctx, "docker", "ps", "-a", "--no-trunc", "--format", "{{.ID}}|{{.Image}}|{{.Names}}|{{.Status}}")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrDockerNotAvailable
		}
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

// GetDockerIDList returns a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	return GetDockerIDListWithContext(context.Background())
}

func GetDockerIDListWithContext(ctx context.Context) ([]string, error) {
	out, err := invoke.CommandWithContext(ctx, "docker", "ps", "-q", "--no-trunc")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrDockerNotAvailable
		}
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

// CgroupCPU returns specified cgroup id CPU status.
// containerID is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerID = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
// on cgroup v1 or base=/sys/fs/cgroup/system.slice/ on cgroup v2.
// The cgroup hierarchy (v1 or v2) is detected automatically through
// /sys/fs/cgroup/cgroup.controllers.
func CgroupCPU(containerID, base string) (*CgroupCPUStat, error) {
	return CgroupCPUWithContext(context.Background(), containerID, base)
}

// CgroupCPUUsage returns specified cgroup id CPU usage.
// containerID is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerID = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
// on cgroup v1 or base=/sys/fs/cgroup/system.slice/ on cgroup v2.
// The cgroup hierarchy (v1 or v2) is detected automatically through
// /sys/fs/cgroup/cgroup.controllers.
func CgroupCPUUsage(containerID, base string) (float64, error) {
	return CgroupCPUUsageWithContext(context.Background(), containerID, base)
}

func CgroupCPUWithContext(ctx context.Context, containerID, base string) (*CgroupCPUStat, error) {
	if isCgroupV2(ctx) {
		return cgroupCPUV2WithContext(ctx, containerID, base)
	}
	statfile, err := getCgroupFilePath(ctx, containerID, base, "cpuacct", "cpuacct.stat")
	if err != nil {
		return nil, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	// empty containerID means all cgroup
	if containerID == "" {
		containerID = "all"
	}

	ret := &CgroupCPUStat{}
	ret.CPU = containerID
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
	usage, err := CgroupCPUUsageWithContext(ctx, containerID, base)
	if err != nil {
		return nil, err
	}
	ret.Usage = usage
	return ret, nil
}

func CgroupCPUUsageWithContext(ctx context.Context, containerID, base string) (float64, error) {
	if isCgroupV2(ctx) {
		stat, err := cgroupCPUV2WithContext(ctx, containerID, base)
		if err != nil {
			return 0.0, err
		}
		return stat.Usage, nil
	}
	usagefile, err := getCgroupFilePath(ctx, containerID, base, "cpuacct", "cpuacct.usage")
	if err != nil {
		return 0.0, err
	}
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

func CgroupCPUDocker(containerID string) (*CgroupCPUStat, error) {
	return CgroupCPUDockerWithContext(context.Background(), containerID)
}

func CgroupCPUUsageDocker(containerID string) (float64, error) {
	return CgroupCPUDockerUsageWithContext(context.Background(), containerID)
}

func CgroupCPUDockerWithContext(ctx context.Context, containerID string) (*CgroupCPUStat, error) {
	// An empty base lets getCgroupFilePath pick the docker directory of the
	// detected cgroup hierarchy (/sys/fs/cgroup/cpuacct/docker on v1).
	return CgroupCPUWithContext(ctx, containerID, "")
}

func CgroupCPUDockerUsageWithContext(ctx context.Context, containerID string) (float64, error) {
	return CgroupCPUUsageWithContext(ctx, containerID, "")
}

// CgroupMem returns specified cgroup id memory status.
// containerID is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerID = docker-<container id>.scope and base=/sys/fs/cgroup/memory/system.slice/
// on cgroup v1 or base=/sys/fs/cgroup/system.slice/ on cgroup v2.
// The cgroup hierarchy (v1 or v2) is detected automatically through
// /sys/fs/cgroup/cgroup.controllers.
func CgroupMem(containerID, base string) (*CgroupMemStat, error) {
	return CgroupMemWithContext(context.Background(), containerID, base)
}

func CgroupMemWithContext(ctx context.Context, containerID, base string) (*CgroupMemStat, error) {
	if isCgroupV2(ctx) {
		return cgroupMemV2WithContext(ctx, containerID, base)
	}
	statfile, err := getCgroupFilePath(ctx, containerID, base, "memory", "memory.stat")
	if err != nil {
		return nil, err
	}

	// empty containerID means all cgroup
	if containerID == "" {
		containerID = "all"
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	ret := &CgroupMemStat{ContainerID: containerID}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "cache":
			ret.Cache = v
		case "rss":
			ret.RSS = v
		case "rssHuge", "rss_huge":
			ret.RSSHuge = v
		case "mappedFile", "mapped_file":
			ret.MappedFile = v
		case "pgpgin":
			ret.Pgpgin = v
		case "pgpgout":
			ret.Pgpgout = v
		case "pgfault":
			ret.Pgfault = v
		case "pgmajfault":
			ret.Pgmajfault = v
		case "inactiveAnon", "inactive_anon":
			ret.InactiveAnon = v
		case "activeAnon", "active_anon":
			ret.ActiveAnon = v
		case "inactiveFile", "inactive_file":
			ret.InactiveFile = v
		case "activeFile", "active_file":
			ret.ActiveFile = v
		case "unevictable":
			ret.Unevictable = v
		case "hierarchicalMemoryLimit", "hierarchical_memory_limit":
			ret.HierarchicalMemoryLimit = v
		case "totalCache", "total_cache":
			ret.TotalCache = v
		case "totalRss", "total_rss":
			ret.TotalRSS = v
		case "totalRssHuge", "total_rss_huge":
			ret.TotalRSSHuge = v
		case "totalMappedFile", "total_mapped_file":
			ret.TotalMappedFile = v
		case "totalPgpgin", "total_pgpgin":
			ret.TotalPgpgIn = v
		case "totalPgpgout", "total_pgpgout":
			ret.TotalPgpgOut = v
		case "totalPgfault", "total_pgfault":
			ret.TotalPgFault = v
		case "totalPgmajfault", "total_pgmajfault":
			ret.TotalPgMajFault = v
		case "totalInactiveAnon", "total_inactive_anon":
			ret.TotalInactiveAnon = v
		case "totalActiveAnon", "total_active_anon":
			ret.TotalActiveAnon = v
		case "totalInactiveFile", "total_inactive_file":
			ret.TotalInactiveFile = v
		case "totalActiveFile", "total_active_file":
			ret.TotalActiveFile = v
		case "totalUnevictable", "total_unevictable":
			ret.TotalUnevictable = v
		}
	}

	r, err := getCgroupMemFile(ctx, containerID, base, "memory.usage_in_bytes")
	if err == nil {
		ret.MemUsageInBytes = r
	}
	r, err = getCgroupMemFile(ctx, containerID, base, "memory.max_usage_in_bytes")
	if err == nil {
		ret.MemMaxUsageInBytes = r
	}
	r, err = getCgroupMemFile(ctx, containerID, base, "memory.limit_in_bytes")
	if err == nil {
		ret.MemLimitInBytes = r
	}
	r, err = getCgroupMemFile(ctx, containerID, base, "memory.failcnt")
	if err == nil {
		ret.MemFailCnt = r
	}

	return ret, nil
}

func CgroupMemDocker(containerID string) (*CgroupMemStat, error) {
	return CgroupMemDockerWithContext(context.Background(), containerID)
}

func CgroupMemDockerWithContext(ctx context.Context, containerID string) (*CgroupMemStat, error) {
	// An empty base lets getCgroupFilePath pick the docker directory of the
	// detected cgroup hierarchy (/sys/fs/cgroup/memory/docker on v1).
	return CgroupMemWithContext(ctx, containerID, "")
}

// cgroupMemV2WithContext reads the memory statistics of a cgroup on the
// cgroup v2 unified hierarchy and maps them onto CgroupMemStat, whose
// fields follow the cgroup v1 naming:
//
//   - memory.current -> MemUsageInBytes
//   - memory.peak -> MemMaxUsageInBytes (Linux 5.19+, stays 0 when the file
//     is absent)
//   - memory.max -> MemLimitInBytes; the literal "max" (no limit) is reported
//     as math.MaxUint64, whereas cgroup v1 reports the kernel sentinel
//     9223372036854771712 for an unlimited cgroup
//   - memory.events "max" -> MemFailCnt, the number of times the usage hit
//     the limit, which is the closest equivalent of the v1 failcnt
//   - memory.stat anon -> RSS, file -> Cache, anon_thp -> RSSHuge,
//     file_mapped -> MappedFile, pgfault -> Pgfault, pgmajfault -> Pgmajfault
//     and inactive_anon, active_anon, inactive_file, active_file, unevictable
//     -> the same-named fields. cgroup v2 counters are hierarchical, so the
//     corresponding Total* fields carry the same values.
//
// Pgpgin, Pgpgout and HierarchicalMemoryLimit have no cgroup v2 equivalent
// and are left at 0.
func cgroupMemV2WithContext(ctx context.Context, containerID, base string) (*CgroupMemStat, error) {
	statfile, err := getCgroupFilePath(ctx, containerID, base, "memory", "memory.stat")
	if err != nil {
		return nil, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}

	ret := &CgroupMemStat{ContainerID: containerID}
	// empty containerID means all cgroup
	if containerID == "" {
		ret.ContainerID = "all"
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "anon":
			ret.RSS, ret.TotalRSS = v, v
		case "file":
			ret.Cache, ret.TotalCache = v, v
		case "anon_thp":
			ret.RSSHuge, ret.TotalRSSHuge = v, v
		case "file_mapped":
			ret.MappedFile, ret.TotalMappedFile = v, v
		case "pgfault":
			ret.Pgfault, ret.TotalPgFault = v, v
		case "pgmajfault":
			ret.Pgmajfault, ret.TotalPgMajFault = v, v
		case "inactive_anon":
			ret.InactiveAnon, ret.TotalInactiveAnon = v, v
		case "active_anon":
			ret.ActiveAnon, ret.TotalActiveAnon = v, v
		case "inactive_file":
			ret.InactiveFile, ret.TotalInactiveFile = v, v
		case "active_file":
			ret.ActiveFile, ret.TotalActiveFile = v, v
		case "unevictable":
			ret.Unevictable, ret.TotalUnevictable = v, v
		}
	}

	r, err := getCgroupMemFileV2(ctx, containerID, base, "memory.current")
	if err == nil {
		ret.MemUsageInBytes = r
	}
	r, err = getCgroupMemFileV2(ctx, containerID, base, "memory.peak")
	if err == nil {
		ret.MemMaxUsageInBytes = r
	}
	r, err = getCgroupMemFileV2(ctx, containerID, base, "memory.max")
	if err == nil {
		ret.MemLimitInBytes = r
	}
	r, err = getCgroupMemFailCntV2(ctx, containerID, base)
	if err == nil {
		ret.MemFailCnt = r
	}

	return ret, nil
}

// cgroupCPUV2WithContext reads the CPU statistics of a cgroup on the cgroup
// v2 unified hierarchy from cpu.stat. usage_usec, user_usec and system_usec
// are converted to seconds, the same unit the cgroup v1 reader produces.
func cgroupCPUV2WithContext(ctx context.Context, containerID, base string) (*CgroupCPUStat, error) {
	statfile, err := getCgroupFilePath(ctx, containerID, base, "cpu", "cpu.stat")
	if err != nil {
		return nil, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	// empty containerID means all cgroup
	if containerID == "" {
		containerID = "all"
	}

	ret := &CgroupCPUStat{}
	ret.CPU = containerID
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		v, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "usage_usec":
			ret.Usage = v / microseconds
		case "user_usec":
			ret.User = v / microseconds
		case "system_usec":
			ret.System = v / microseconds
		}
	}
	return ret, nil
}

// isCgroupV2 reports whether the host mounts the cgroup v2 unified hierarchy
// at /sys/fs/cgroup. The hybrid layout (v1 controllers plus a "unified"
// subtree) has no cgroup.controllers file at the root and is treated as v1.
// The result is deliberately not cached so that HOST_SYS and per-context
// environment overrides are honoured on every call.
func isCgroupV2(ctx context.Context) bool {
	return common.PathExists(common.HostSysWithContext(ctx, "fs/cgroup/cgroup.controllers"))
}

// cgroupDir returns the host path of dir inside the cgroup tree. On cgroup v2
// there is a single unified tree (/sys/fs/cgroup/<dir>), on cgroup v1 each
// controller has its own tree (/sys/fs/cgroup/<target>/<dir>).
func cgroupDir(ctx context.Context, v2 bool, target, dir string) string {
	if v2 {
		return common.HostSysWithContext(ctx, "fs/cgroup", dir)
	}
	return common.HostSysWithContext(ctx, "fs/cgroup", target, dir)
}

// getCgroupFilePath constructs file path to get targeted stats file.
// When base is empty the docker directory of the detected cgroup hierarchy
// is used (cgroupfs driver layout). When the container directory does not
// exist under base, the systemd driver layout
// (system.slice/docker-<containerID>.scope) is tried instead.
func getCgroupFilePath(ctx context.Context, containerID, base, target, file string) (string, error) {
	// Prevent a caller-supplied containerID from escaping the cgroup base
	// directory via path separators or ".." traversal.
	if strings.ContainsAny(containerID, `/\`) || strings.Contains(containerID, "..") {
		return "", fmt.Errorf("invalid container ID %q", containerID)
	}
	v2 := isCgroupV2(ctx)
	if base == "" {
		base = cgroupDir(ctx, v2, target, "docker")
	}
	statfile := path.Join(base, containerID, file)

	if _, err := os.Stat(statfile); os.IsNotExist(err) {
		statfile = path.Join(cgroupDir(ctx, v2, target, "system.slice"), "docker-"+containerID+".scope", file)
	}

	return statfile, nil
}

// getCgroupMemFileV2 reads a single-value cgroup v2 memory file and returns
// the contents as uint64. The literal "max" (no limit) is returned as
// math.MaxUint64.
func getCgroupMemFileV2(ctx context.Context, containerID, base, file string) (uint64, error) {
	statfile, err := getCgroupFilePath(ctx, containerID, base, "memory", file)
	if err != nil {
		return 0, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return 0, err
	}
	if len(lines) != 1 {
		return 0, fmt.Errorf("wrong format file: %s", statfile)
	}
	value := strings.TrimSpace(lines[0])
	if value == "max" {
		return math.MaxUint64, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

// getCgroupMemFailCntV2 returns the "max" counter of memory.events, i.e. the
// number of times the memory usage of the cgroup hit its limit.
func getCgroupMemFailCntV2(ctx context.Context, containerID, base string) (uint64, error) {
	statfile, err := getCgroupFilePath(ctx, containerID, base, "memory", "memory.events")
	if err != nil {
		return 0, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return 0, err
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "max" {
			return strconv.ParseUint(fields[1], 10, 64)
		}
	}
	return 0, fmt.Errorf("no max event in %s", statfile)
}

// getCgroupMemFile reads a cgroup file and return the contents as uint64.
func getCgroupMemFile(ctx context.Context, containerID, base, file string) (uint64, error) {
	statfile, err := getCgroupFilePath(ctx, containerID, base, "memory", file)
	if err != nil {
		return 0, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return 0, err
	}
	if len(lines) != 1 {
		return 0, fmt.Errorf("wrong format file: %s", statfile)
	}
	return strconv.ParseUint(lines[0], 10, 64)
}
