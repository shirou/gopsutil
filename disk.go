package gopsutil

type Disk struct{}

type Disk_usage struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

type Disk_partition struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Fstype     string `json:"fstype"`
	Opts       string `json:"opts"`
}

type Disk_IO_Counters struct {
	ReadCount  uint64 `json:"readCount"`
	WriteCount uint64 `json:"writeCount"`
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
	ReadTime   uint64 `json:"readTime"`
	WriteTime  uint64 `json:"writeTime"`
}

func NewDisk() Disk {
	d := Disk{}
	return d
}
