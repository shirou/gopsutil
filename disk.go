package gopsutil

import (
	"encoding/json"
)

type DiskUsageStat struct {
	Path              string  `json:"path"`
	Total             uint64  `json:"total"`
	Free              uint64  `json:"free"`
	Used              uint64  `json:"used"`
	UsedPercent       float64 `json:"usedPercent"`
	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

type DiskPartitionStat struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Fstype     string `json:"fstype"`
	Opts       string `json:"opts"`
}

type DiskIOCountersStat struct {
	ReadCount  uint64 `json:"readCount"`
	WriteCount uint64 `json:"writeCount"`
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
	ReadTime   uint64 `json:"readTime"`
	WriteTime  uint64 `json:"writeTime"`
	Name       string `json:"name"`
	IoTime     uint64 `json:"ioTime"`
}

func (d DiskUsageStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

func (d DiskPartitionStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

func (d DiskIOCountersStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}
