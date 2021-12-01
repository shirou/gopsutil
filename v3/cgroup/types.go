package cgroup

import (
	"encoding/json"

	"github.com/shirou/gopsutil/v3/cpu"
)

const nanoseconds = 1e9

type CgroupCPUStat struct {
	cpu.TimesStat
	Usage float64
}

type CgroupMemStat struct {
	// Total amount of RAM on this system
	Total uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used uint64 `json:"used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	UsedPercent float64 `json:"usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `json:"free"`

	// OS X / BSD specific numbers:
	// http://www.macyourself.com/2010/02/17/what-is-free-wired-active-and-inactive-system-memory-ram/
	Active   uint64 `json:"active"`
	Inactive uint64 `json:"inactive"`

	ContainerID             string `json:"containerID"`
	Cache                   uint64 `json:"cache"`
	RSS                     uint64 `json:"rss"`
	RSSHuge                 uint64 `json:"rssHuge"`
	MappedFile              uint64 `json:"mappedFile"`
	Pgpgin                  uint64 `json:"pgpgin"`
	Pgpgout                 uint64 `json:"pgpgout"`
	Pgfault                 uint64 `json:"pgfault"`
	Pgmajfault              uint64 `json:"pgmajfault"`
	InactiveAnon            uint64 `json:"inactiveAnon"`
	ActiveAnon              uint64 `json:"activeAnon"`
	InactiveFile            uint64 `json:"inactiveFile"`
	ActiveFile              uint64 `json:"activeFile"`
	Unevictable             uint64 `json:"unevictable"`
	HierarchicalMemoryLimit uint64 `json:"hierarchicalMemoryLimit"`
	TotalCache              uint64 `json:"totalCache"`
	TotalRSS                uint64 `json:"totalRss"`
	TotalRSSHuge            uint64 `json:"totalRssHuge"`
	TotalMappedFile         uint64 `json:"totalMappedFile"`
	TotalPgpgIn             uint64 `json:"totalPgpgin"`
	TotalPgpgOut            uint64 `json:"totalPgpgout"`
	TotalPgFault            uint64 `json:"totalPgfault"`
	TotalPgMajFault         uint64 `json:"totalPgmajfault"`
	TotalInactiveAnon       uint64 `json:"totalInactiveAnon"`
	TotalActiveAnon         uint64 `json:"totalActiveAnon"`
	TotalInactiveFile       uint64 `json:"totalInactiveFile"`
	TotalActiveFile         uint64 `json:"totalActiveFile"`
	TotalUnevictable        uint64 `json:"totalUnevictable"`
	MemUsageInBytes         uint64 `json:"memUsageInBytes"`
	MemMaxUsageInBytes      uint64 `json:"memMaxUsageInBytes"`
	MemLimitInBytes         uint64 `json:"memoryLimitInBytes"`
	MemFailCnt              uint64 `json:"memoryFailcnt"`
	Swap                    uint64 `json:"swap"`
}

func (m CgroupMemStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}