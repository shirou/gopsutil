package gopsutil

type Process struct {
	Pid         int32  `json:"pid"`
}

type Open_filesStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

type Memory_infoStat struct {
	RSS uint64 `json:"rss"` // bytes
	VMS uint64 `json:"vms"` // bytes
}

type RlimitStat struct {
	Resource int32 `json:"resource"`
	Soft     int32 `json:"soft"`
	Hard     int32 `json:"hard"`
}

type Io_countersStat struct {
	Read_count  int32 `json:"read_count"`
	Write_count int32 `json:"write_count"`
	Read_bytes  int32 `json:"read_bytes"`
	Write_bytes int32 `json:"write_bytes"`
}

func Pid_exists(pid int32) (bool, error) {
	pids, err := Pids()
	if err != nil {
		return false, err
	}

	for _, i := range pids {
		if i == pid {
			return true, err
		}
	}

	return false, err
}
