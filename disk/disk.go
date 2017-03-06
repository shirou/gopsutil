package disk

import (
	"encoding/json"

	"github.com/shirou/gopsutil/internal/common"
)

var invoke common.Invoker

func init() {
	invoke = common.Invoke{}
}

type UsageStat struct {
	Path              string  `json:"path" bson:"path"`
	Fstype            string  `json:"fstype" bson:"fstype"`
	Total             uint64  `json:"total" bson:"total"`
	Free              uint64  `json:"free" bson:"free"`
	Used              uint64  `json:"used" bson:"used"`
	UsedPercent       float64 `json:"usedPercent" bson:"usedPercent"`
	InodesTotal       uint64  `json:"inodesTotal" bson:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed" bson:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree" bson:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent" bson:"inodesUsedPercent"`
}

type PartitionStat struct {
	Device     string `json:"device" bson:"device"`
	Mountpoint string `json:"mountpoint" bson:"mountpoint"`
	Fstype     string `json:"fstype" bson:"fstype"`
	Opts       string `json:"opts" bson:"opts"`
}

type IOCountersStat struct {
	ReadCount        uint64 `json:"readCount" bson:"readCount"`
	MergedReadCount  uint64 `json:"mergedReadCount" bson:"mergedReadCount"`
	WriteCount       uint64 `json:"writeCount" bson:"writeCount"`
	MergedWriteCount uint64 `json:"mergedWriteCount" bson:"mergedWriteCount"`
	ReadBytes        uint64 `json:"readBytes" bson:"readBytes"`
	WriteBytes       uint64 `json:"writeBytes" bson:"writeBytes"`
	ReadTime         uint64 `json:"readTime" bson:"readTime"`
	WriteTime        uint64 `json:"writeTime" bson:"writeTime"`
	IopsInProgress   uint64 `json:"iopsInProgress" bson:"iopsInProgress"`
	IoTime           uint64 `json:"ioTime" bson:"ioTime"`
	WeightedIO       uint64 `json:"weightedIO" bson:"weightedIO"`
	Name             string `json:"name" bson:"name"`
	SerialNumber     string `json:"serialNumber" bson:"serialNumber"`
}

func (d UsageStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

func (d PartitionStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

func (d IOCountersStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}
