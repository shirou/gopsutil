package docker

import (
	"errors"

	"github.com/shirou/gopsutil/internal/common"
)

var ErrDockerNotAvailable = errors.New("docker not available")
var ErrCgroupNotAvailable = errors.New("cgroup not available")

var invoke common.Invoker

func init() {
	invoke = common.Invoke{}
}

type CgroupMemStat struct {
	ContainerID             string `json:"containerID" bson:"containerID"`
	Cache                   uint64 `json:"cache" bson:"cache"`
	RSS                     uint64 `json:"rss" bson:"rss"`
	RSSHuge                 uint64 `json:"rssHuge" bson:"rssHuge"`
	MappedFile              uint64 `json:"mappedFile" bson:"mappedFile"`
	Pgpgin                  uint64 `json:"pgpgin" bson:"pgpgin"`
	Pgpgout                 uint64 `json:"pgpgout" bson:"pgpgout"`
	Pgfault                 uint64 `json:"pgfault" bson:"pgfault"`
	Pgmajfault              uint64 `json:"pgmajfault" bson:"pgmajfault"`
	InactiveAnon            uint64 `json:"inactiveAnon" bson:"inactiveAnon"`
	ActiveAnon              uint64 `json:"activeAnon" bson:"activeAnon"`
	InactiveFile            uint64 `json:"inactiveFile" bson:"inactiveFile"`
	ActiveFile              uint64 `json:"activeFile" bson:"activeFile"`
	Unevictable             uint64 `json:"unevictable" bson:"unevictable"`
	HierarchicalMemoryLimit uint64 `json:"hierarchicalMemoryLimit" bson:"hierarchicalMemoryLimit"`
	TotalCache              uint64 `json:"totalCache" bson:"totalCache"`
	TotalRSS                uint64 `json:"totalRss" bson:"totalRss"`
	TotalRSSHuge            uint64 `json:"totalRssHuge" bson:"totalRssHuge"`
	TotalMappedFile         uint64 `json:"totalMappedFile" bson:"totalMappedFile"`
	TotalPgpgIn             uint64 `json:"totalPgpgin" bson:"totalPgpgin"`
	TotalPgpgOut            uint64 `json:"totalPgpgout" bson:"totalPgpgout"`
	TotalPgFault            uint64 `json:"totalPgfault" bson:"totalPgfault"`
	TotalPgMajFault         uint64 `json:"totalPgmajfault" bson:"totalPgmajfault"`
	TotalInactiveAnon       uint64 `json:"totalInactiveAnon" bson:"totalInactiveAnon"`
	TotalActiveAnon         uint64 `json:"totalActiveAnon" bson:"totalActiveAnon"`
	TotalInactiveFile       uint64 `json:"totalInactiveFile" bson:"totalInactiveFile"`
	TotalActiveFile         uint64 `json:"totalActiveFile" bson:"totalActiveFile"`
	TotalUnevictable        uint64 `json:"totalUnevictable" bson:"totalUnevictable"`
	MemUsageInBytes         uint64 `json:"memUsageInBytes" bson:"memUsageInBytes"`
	MemMaxUsageInBytes      uint64 `json:"memMaxUsageInBytes" bson:"memMaxUsageInBytes"`
	MemLimitInBytes         uint64 `json:"memoryLimitInBbytes" bson:"memoryLimitInBbytes"`
	MemFailCnt              uint64 `json:"memoryFailcnt" bson:"memoryFailcnt"`
}

type CgroupDockerStat struct {
	ContainerID string `json:"containerID" bson:"containerID"`
	Name        string `json:"name" bson:"name"`
	Image       string `json:"image" bson:"image"`
	Status      string `json:"status" bson:"status"`
	Running     bool   `json:"running" bson:"running"`
}
