package gopsutil

type Process struct {
	Pid         int32  `json:"pid"`
	Ppid        int32  `json:"ppid"`
	Name        string `json:"name"`
	Exe         string `json:"exe"`
	Cmdline     string `json:"cmdline"`
	Create_time int64  `json:"create_time"`
	//	Parent           Process // FIXME: recursive
	Status           string          `json:"status"`
	Cwd              string          `json:"cwd"`
	Username         string          `json:"username"`
	Uids             []int32         `json:"uids"`
	Gids             []int32         `json:"gids"`
	Terminal         string          `json:"terminal"`
	Nice             int32           `json:"nice"`
	Ionice           int32           `json:"ionice"`
	Rlimit           []RlimitStat    `json:"rlimit"`
	Io_counters      Io_countersStat `json:"io_counter"`
	Num_ctx_switches int32           `json:"num_ctx_switches"`
	Num_fds          int32           `json:"num_fds"`
	Num_handles      int32           `json:"num_handles"`
	Num_Threads      int32           `json:"nunm_threads"`
	//	Threads  map[string]string `json:"threads"`
	Cpu_times CPU_TimesStat `json:"cpu_times"`
	//	Cpu_percent `json:"cpu_percent"`
	Cpu_affinity   []int32           `json:"cpu_affinity"`
	Memory_info    Memory_infoStat   `json:"memory_info"`
	Memory_info_ex map[string]string `json:"memori_info_ex"`
	Memory_percent float32           `json:"memory_percent"`
	Memory_maps    []Memory_mapsStat `json:"memory_maps"`
	//	Children       []Process // FIXME: recursive `json:"children"`
	Open_files  []Open_filesStat     `json:"open_files"`
	Connections []Net_connectionStat `json:"connections"`
	Is_running  bool                 `json:"is_running"`
}

type Open_filesStat struct {
	Path string `json:"path"`
	Fd   uint32 `json:"fd"`
}

type Memory_infoStat struct {
	RSS int32 `json:"rss"` // bytes
	VMS int32 `json:"vms"` // bytes
}

type Memory_mapsStat struct {
	Path      string `json:"path"`
	RSS       int32  `json:"rss"`
	Anonymous int32  `json:"anonymous"`
	Swap      int32  `json:"swap"`
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
