// +build linux

package gopsutil

import (
	"encoding/json"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type CgroupMemStat struct {
	ContainerID               string `json:"containerid"`
	Cache                     uint64 `json:"cache"`
	RSS                       uint64 `json:"rss"`
	Rss_huge                  uint64 `json:"rss_huge"`
	Mapped_file               uint64 `json:"mapped_file"`
	Pgpgin                    uint64 `json:"pgpgin"`
	Pgpgout                   uint64 `json:"pgpgout"`
	Pgfault                   uint64 `json:"pgfault"`
	Pgmajfault                uint64 `json:"pgmajfault"`
	Inactive_anon             uint64 `json:"inactive_anon"`
	Active_anon               uint64 `json:"active_anon"`
	Inactive_file             uint64 `json:"inactive_file"`
	Active_file               uint64 `json:"active_file"`
	Unevictable               uint64 `json:"unevictable"`
	Hierarchical_memory_limit uint64 `json:"hierarchical_memory_limit"`
	Total_cache               uint64 `json:"total_cache"`
	Total_rss                 uint64 `json:"total_rss"`
	Total_rss_huge            uint64 `json:"total_rss_huge"`
	Total_mapped_file         uint64 `json:"total_mapped_file"`
	Total_pgpgin              uint64 `json:"total_pgpgin"`
	Total_pgpgout             uint64 `json:"total_pgpgout"`
	Total_pgfault             uint64 `json:"total_pgfault"`
	Total_pgmajfault          uint64 `json:"total_pgmajfault"`
	Total_inactive_anon       uint64 `json:"total_inactive_anon"`
	Total_active_anon         uint64 `json:"total_active_anon"`
	Total_inactive_file       uint64 `json:"total_inactive_file"`
	Total_active_file         uint64 `json:"total_active_file"`
	Total_unevictable         uint64 `json:"total_unevictable"`
}

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	out, err := exec.Command("docker", "ps", "-q", "--no-trunc").Output()
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(out), "\n")
	ret := make([]string, 0, len(lines))

	for _, l := range lines {
		ret = append(ret, l)
	}

	return ret, nil
}

// CgroupCPU returnes specified cgroup id CPU status.
// containerid is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerid = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerid string, base string) (*CPUTimesStat, error) {
	if len(base) == 0 {
		base = "/sys/fs/cgroup/cpuacct/docker"
	}
	path := path.Join(base, containerid, "cpuacct.stat")

	lines, _ := readLines(path)
	// empty containerid means all cgroup
	if len(containerid) == 0 {
		containerid = "all"
	}
	ret := &CPUTimesStat{CPU: containerid}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if fields[0] == "user" {
			user, err := strconv.ParseFloat(fields[1], 32)
			if err == nil {
				ret.User = float32(user)
			}
		}
		if fields[0] == "system" {
			system, err := strconv.ParseFloat(fields[1], 32)
			if err == nil {
				ret.System = float32(system)
			}
		}
	}

	return ret, nil
}

func CgroupCPUDocker(containerid string) (*CPUTimesStat, error) {
	return CgroupCPU(containerid, "/sys/fs/cgroup/cpuacct/docker")
}

func CgroupMem(containerid string, base string) (*CgroupMemStat, error) {
	if len(base) == 0 {
		base = "/sys/fs/cgroup/memory/docker"
	}
	path := path.Join(base, containerid, "memory.stat")
	// empty containerid means all cgroup
	if len(containerid) == 0 {
		containerid = "all"
	}
	lines, _ := readLines(path)
	ret := &CgroupMemStat{ContainerID: containerid}
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
		case "rss_huge":
			ret.Rss_huge = v
		case "mapped_file":
			ret.Mapped_file = v
		case "pgpgin":
			ret.Pgpgin = v
		case "pgpgout":
			ret.Pgpgout = v
		case "pgfault":
			ret.Pgfault = v
		case "pgmajfault":
			ret.Pgmajfault = v
		case "inactive_anon":
			ret.Inactive_anon = v
		case "active_anon":
			ret.Active_anon = v
		case "inactive_file":
			ret.Inactive_file = v
		case "active_file":
			ret.Active_file = v
		case "unevictable":
			ret.Unevictable = v
		case "hierarchical_memory_limit":
			ret.Hierarchical_memory_limit = v
		case "total_cache":
			ret.Total_cache = v
		case "total_rss":
			ret.Total_rss = v
		case "total_rss_huge":
			ret.Total_rss_huge = v
		case "total_mapped_file":
			ret.Total_mapped_file = v
		case "total_pgpgin":
			ret.Total_pgpgin = v
		case "total_pgpgout":
			ret.Total_pgpgout = v
		case "total_pgfault":
			ret.Total_pgfault = v
		case "total_pgmajfault":
			ret.Total_pgmajfault = v
		case "total_inactive_anon":
			ret.Total_inactive_anon = v
		case "total_active_anon":
			ret.Total_active_anon = v
		case "total_inactive_file":
			ret.Total_inactive_file = v
		case "total_active_file":
			ret.Total_active_file = v
		case "total_unevictable":
			ret.Total_unevictable = v
		}
	}
	return ret, nil
}

func CgroupMemDocker(containerid string) (*CgroupMemStat, error) {
	return CgroupMem(containerid, "/sys/fs/cgroup/memory/docker")
}

func (m CgroupMemStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
