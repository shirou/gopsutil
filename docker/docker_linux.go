// +build linux

package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	cpu "github.com/DataDog/gopsutil/cpu"
	"github.com/DataDog/gopsutil/internal/common"
)

// GetDockerStat returns a list of Docker basic stats.
// This requires certain permission.
func GetDockerStat() ([]CgroupDockerStat, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, ErrDockerNotAvailable
	}

	out, err := invoke.Command(path, "ps", "-a", "--no-trunc", "--format", "{{.ID}}|{{.Image}}|{{.Names}}|{{.Status}}")
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

// Generates a mapping of PIDs to container metadata.
func GetContainerStatsByPID() (map[int32]ContainerStat, error) {
	containerMap := make(map[int32]ContainerStat)
	path, err := cgroupMountPoint("cpuacct")
	if err != nil {
		return nil, err
	}
	if common.PathExists(path) {
		contents, err := common.ListDirectory(path)
		if err != nil {
			return nil, err
		}

		// If docker containers exist, collect their stats.
		if common.StringsContains(contents, "docker") {
			dockerStats, err := GetDockerStat()
			if err == nil {
				for _, dockerStat := range dockerStats {
					if !dockerStat.Running {
						continue
					}

					dockerPids, err := CgroupPIDsDocker(dockerStat.ContainerID)
					if err != nil {
						continue
					}
					memstat, err := CgroupMemDocker(dockerStat.ContainerID)
					if err != nil {
						continue
					}
					cpuLimit, err := CgroupCPULimitDocker(dockerStat.ContainerID)
					if err != nil {
						continue
					}

					containerStat := ContainerStat{
						Type:     "Docker",
						Name:     dockerStat.Name,
						ID:       dockerStat.ContainerID,
						Image:    dockerStat.Image,
						MemStat:  memstat,
						CPULimit: cpuLimit,
					}

					for _, pid := range dockerPids {
						containerMap[pid] = containerStat
					}
				}
			}
		}
	}

	return containerMap, nil
}

func (c CgroupDockerStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, ErrDockerNotAvailable
	}

	out, err := invoke.Command(path, "ps", "-q", "--no-trunc")
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

// CgroupCPU returnes specified cgroup id CPU status.
// containerID is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerID = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerID string, base string) (*cpu.TimesStat, error) {
	statfile, err := getCgroupFilePath(containerID, base, "cpuacct", "cpuacct.stat")
	if err != nil {
		return nil, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	// empty containerID means all cgroup
	if len(containerID) == 0 {
		containerID = "all"
	}
	ret := &cpu.TimesStat{CPU: containerID}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if fields[0] == "user" {
			user, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.User = float64(user)
			}
		}
		if fields[0] == "system" {
			system, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.System = float64(system)
			}
		}
	}

	return ret, nil
}

// CgroupCPULimit would show CPU limit for each container
// it does so by checking the cpu period and cpu quota config
// if a user does this:
//
//	docker run --cpus='0.5' ubuntu:latest
//
// we should return 50% for that container
func CgroupCPULimit(containerID string, base string) (float64, error) {
	periodFile, err := getCgroupFilePath(containerID, base, "cpu", "cpu.cfs_period_us")
	if err != nil {
		return 0.0, err
	}
	quotaFile, err := getCgroupFilePath(containerID, base, "cpu", "cpu.cfs_quota_us")
	if err != nil {
		return 0.0, err
	}
	plines, err := common.ReadLines(periodFile)
	if err != nil {
		return 0.0, err
	}
	qlines, err := common.ReadLines(quotaFile)
	if err != nil {
		return 0.0, err
	}
	period, err := strconv.ParseFloat(plines[0], 64)
	if err != nil {
		return 0.0, err
	}
	quota, err := strconv.ParseFloat(qlines[0], 64)
	if err != nil {
		return 0.0, err
	}
	// default cpu limit is 100%
	limit := 100.0
	if (period > 0) && (quota > 0) {
		limit = (quota / period) * 100.0
	}
	return limit, nil
}

func CgroupCPULimitDocker(containerID string) (float64, error) {
	p, err := cgroupMountPoint("cpu")
	if err != nil {
		return 0.0, err
	}
	return CgroupCPULimit(containerID, filepath.Join(p, "docker"))
}

func CgroupCPUDocker(containerID string) (*cpu.TimesStat, error) {
	p, err := cgroupMountPoint("cpuacct")
	if err != nil {
		return nil, err
	}
	return CgroupCPU(containerID, filepath.Join(p, "docker"))
}

// CgroupPIDs retrieves the PIDs running within a given container.
func CgroupPIDs(containerID string, base string) ([]int32, error) {
	statfile, err := getCgroupFilePath(containerID, base, "cpuacct", "cgroup.procs")
	if err != nil {
		return nil, err
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}

	pids := make([]int32, 0, len(lines))
	for _, line := range lines {
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, int32(pid))
		}
	}

	return pids, nil
}

func CgroupPIDsDocker(containerID string) ([]int32, error) {
	p, err := cgroupMountPoint("cpuacct")
	if err != nil {
		return []int32{}, err
	}
	return CgroupPIDs(containerID, filepath.Join(p, "docker"))
}

func CgroupMem(containerID string, base string) (*CgroupMemStat, error) {
	statfile, err := getCgroupFilePath(containerID, base, "memory", "memory.stat")
	if err != nil {
		return nil, err
	}

	// empty containerID means all cgroup
	if len(containerID) == 0 {
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
		case "rssHuge":
			ret.RSSHuge = v
		case "mappedFile":
			ret.MappedFile = v
		case "pgpgin":
			ret.Pgpgin = v
		case "pgpgout":
			ret.Pgpgout = v
		case "pgfault":
			ret.Pgfault = v
		case "pgmajfault":
			ret.Pgmajfault = v
		case "inactiveAnon":
			ret.InactiveAnon = v
		case "activeAnon":
			ret.ActiveAnon = v
		case "inactiveFile":
			ret.InactiveFile = v
		case "activeFile":
			ret.ActiveFile = v
		case "unevictable":
			ret.Unevictable = v
		case "hierarchicalMemoryLimit":
			ret.HierarchicalMemoryLimit = v
		case "totalCache":
			ret.TotalCache = v
		case "totalRss":
			ret.TotalRSS = v
		case "totalRssHuge":
			ret.TotalRSSHuge = v
		case "totalMappedFile":
			ret.TotalMappedFile = v
		case "totalPgpgin":
			ret.TotalPgpgIn = v
		case "totalPgpgout":
			ret.TotalPgpgOut = v
		case "totalPgfault":
			ret.TotalPgFault = v
		case "totalPgmajfault":
			ret.TotalPgMajFault = v
		case "totalInactiveAnon":
			ret.TotalInactiveAnon = v
		case "totalActiveAnon":
			ret.TotalActiveAnon = v
		case "totalInactiveFile":
			ret.TotalInactiveFile = v
		case "totalActiveFile":
			ret.TotalActiveFile = v
		case "totalUnevictable":
			ret.TotalUnevictable = v
		}
	}

	r, err := getCgroupMemFile(containerID, base, "memory.usage_in_bytes")
	if err == nil {
		ret.MemUsageInBytes = r
	}
	r, err = getCgroupMemFile(containerID, base, "memory.max_usage_in_bytes")
	if err == nil {
		ret.MemMaxUsageInBytes = r
	}
	r, err = getCgroupMemFile(containerID, base, "memory.limit_in_bytes")
	if err == nil {
		ret.MemLimitInBytes = r
	}
	r, err = getCgroupMemFile(containerID, base, "memoryFailcnt")
	if err == nil {
		ret.MemFailCnt = r
	}

	return ret, nil
}

func CgroupMemDocker(containerID string) (*CgroupMemStat, error) {
	p, err := cgroupMountPoint("memory")
	if err != nil {
		return nil, err
	}
	return CgroupMem(containerID, filepath.Join(p, "docker"))
}

func (m CgroupMemStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// getCgroupFilePath constructs file path to get targetted stats file.
func getCgroupFilePath(containerID, base, target, file string) (string, error) {
	if len(base) == 0 {
		base, _ = cgroupMountPoint(target)
		base = filepath.Join(base, "docker")
	}
	statfile := filepath.Join(base, containerID, file)

	if _, err := os.Stat(statfile); os.IsNotExist(err) {
		base, err = cgroupMountPoint(target)
		if err != nil {
			return "", err
		}
		statfile = filepath.Join(base, "system.slice", fmt.Sprintf("docker-%s.scope", containerID), file)
	}

	return statfile, nil
}

// getCgroupMemFile reads a cgroup file and return the contents as uint64.
func getCgroupMemFile(containerID, base, file string) (uint64, error) {
	statfile, err := getCgroupFilePath(containerID, base, "memory", file)
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
	v, err := strconv.ParseUint(lines[0], 10, 64)
	if err != nil {
		return 0, err
	}
	// limit_in_bytes is a special case here, it's possible that it shows a rediculous number,
	// in which case it represents unlimited, so return 0 here
	if (file == "memory.limit_in_bytes") && (v > uint64(math.Pow(2, 60))) {
		v = 0
	}
	return v, nil
}

// function to get the mount point of cgroup. by default it should be under /sys/fs/cgroup but
// it could be mounted anywhere else if manually defined. Example cgroup entries in /proc/mounts would be
//	 cgroup /sys/fs/cgroup/cpuset cgroup rw,relatime,cpuset 0 0
//	 cgroup /sys/fs/cgroup/cpu cgroup rw,relatime,cpu 0 0
//	 cgroup /sys/fs/cgroup/cpuacct cgroup rw,relatime,cpuacct 0 0
//	 cgroup /sys/fs/cgroup/memory cgroup rw,relatime,memory 0 0
//	 cgroup /sys/fs/cgroup/devices cgroup rw,relatime,devices 0 0
//	 cgroup /sys/fs/cgroup/freezer cgroup rw,relatime,freezer 0 0
//	 cgroup /sys/fs/cgroup/blkio cgroup rw,relatime,blkio 0 0
//	 cgroup /sys/fs/cgroup/perf_event cgroup rw,relatime,perf_event 0 0
//	 cgroup /sys/fs/cgroup/hugetlb cgroup rw,relatime,hugetlb 0 0
// examples of target would be:
// blkio  cpu  cpuacct  cpuset  devices  freezer  hugetlb  memory  net_cls  net_prio  perf_event  pids  systemd
func cgroupMountPoint(target string) (string, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cgroups := []string{}
	// get all cgroup entries
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "cgroup ") {
			cgroups = append(cgroups, text)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	// old cgroup style
	if len(cgroups) == 1 {
		tokens := strings.Split(cgroups[0], " ")
		return tokens[1], nil
	}

	var candidate string
	for _, cgroup := range cgroups {
		tokens := strings.Split(cgroup, " ")
		// see if the target is the suffix of the mount directory
		if strings.HasSuffix(tokens[1], target) {
			candidate = tokens[1]
		}
	}
	if candidate == "" {
		return candidate, fmt.Errorf("Mount point for cgroup %s is not found!", target)
	}
	return candidate, nil
}
