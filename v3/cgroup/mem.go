package cgroup

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func CgroupMem() (*CgroupMemStat, error) {
	return CgroupMemWithContext(context.Background(), getCgroupFilePath("", "memory", "memory.stat"))
}

func asUInt(value string) uint64 {
	v, _ := strconv.ParseUint(value, 10, 64)
	return v
}

func CgroupMemWithContext(ctx context.Context, statfile string) (*CgroupMemStat, error) {
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	ret := &CgroupMemStat{}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		switch fields[0] {
		case "docker":
			ret.ContainerID = fields[1]
		case "cache":
			ret.Cache = asUInt(fields[1])
		case "rss":
			ret.RSS = asUInt(fields[1])
		case "rssHuge":
			ret.RSSHuge = asUInt(fields[1])
		case "mappedFile":
			ret.MappedFile = asUInt(fields[1])
		case "pgpgin":
			ret.Pgpgin = asUInt(fields[1])
		case "pgpgout":
			ret.Pgpgout = asUInt(fields[1])
		case "pgfault":
			ret.Pgfault = asUInt(fields[1])
		case "pgmajfault":
			ret.Pgmajfault = asUInt(fields[1])
		case "inactiveAnon", "inactive_anon":
			ret.InactiveAnon = asUInt(fields[1])
		case "activeAnon", "active_anon":
			ret.ActiveAnon = asUInt(fields[1])
		case "inactiveFile", "inactive_file":
			ret.InactiveFile = asUInt(fields[1])
		case "activeFile", "active_file":
			ret.ActiveFile = asUInt(fields[1])
		case "unevictable":
			ret.Unevictable = asUInt(fields[1])
		case "hierarchicalMemoryLimit", "hierarchical_memory_limit":
			ret.HierarchicalMemoryLimit = asUInt(fields[1])
		case "totalCache", "total_cache":
			ret.TotalCache = asUInt(fields[1])
		case "totalRss", "total_rss":
			ret.TotalRSS = asUInt(fields[1])
		case "totalRssHuge", "total_rss_huge":
			ret.TotalRSSHuge = asUInt(fields[1])
		case "totalMappedFile", "total_mapped_file":
			ret.TotalMappedFile = asUInt(fields[1])
		case "totalPgpgin", "total_pgpgin":
			ret.TotalPgpgIn = asUInt(fields[1])
		case "totalPgpgout", "total_pgpgout":
			ret.TotalPgpgOut = asUInt(fields[1])
		case "totalPgfault", "total_pgfault":
			ret.TotalPgFault = asUInt(fields[1])
		case "totalPgmajfault", "total_pgmajfault":
			ret.TotalPgMajFault = asUInt(fields[1])
		case "totalInactiveAnon", "total_inactive_anon":
			ret.TotalInactiveAnon = asUInt(fields[1])
		case "totalActiveAnon", "total_active_anon":
			ret.TotalActiveAnon = asUInt(fields[1])
		case "totalInactiveFile", "total_inactive_file":
			ret.TotalInactiveFile = asUInt(fields[1])
		case "totalActiveFile", "total_active_file":
			ret.TotalActiveFile = asUInt(fields[1])
		case "totalUnevictable", "total_unevictable":
			ret.TotalUnevictable = asUInt(fields[1])
		default:
			fmt.Println(fields[0], " = ", fields[1])
		}
	}

	r, err := ioutil.ReadFile(path.Join(path.Dir(statfile), "memory.usage_in_bytes"))
	if err == nil {
		ret.MemUsageInBytes = asUInt(string(r))
	}
	r, err = ioutil.ReadFile(path.Join(path.Dir(statfile), "memory.max_usage_in_bytes"))
	if err == nil {
		ret.MemMaxUsageInBytes = asUInt(string(r))
	}
	r, err = ioutil.ReadFile(path.Join(path.Dir(statfile), "memory.limit_in_bytes"))
	if err == nil {
		ret.MemLimitInBytes = asUInt(string(r))
	}
	r, err = ioutil.ReadFile(path.Join(path.Dir(statfile), "memory.failcnt"))
	if err == nil {
		ret.MemFailCnt = asUInt(string(r))
	}

	return ret, nil
}