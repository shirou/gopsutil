package gopsutil

type Process struct {
	Pid int32 `json:"pid"`
}

type OpenFilesStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

type MemoryInfoStat struct {
	RSS uint64 `json:"rss"` // bytes
	VMS uint64 `json:"vms"` // bytes
}

type RlimitStat struct {
	Resource int32 `json:"resource"`
	Soft     int32 `json:"soft"`
	Hard     int32 `json:"hard"`
}

type IoCountersStat struct {
	ReadCount  int32 `json:"read_count"`
	WriteCount int32 `json:"write_count"`
	ReadBytes  int32 `json:"read_bytes"`
	WriteBytes int32 `json:"write_bytes"`
}

func PidExists(pid int32) (bool, error) {
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
