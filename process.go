package gopsutil

type Process struct {
	Pid         int32  `json:"pid"`
	Ppid        int32  `json:"ppid"`
	Name        string `json:"name"`
	Exe         string `json:"exe"`
	Cmdline     string `json:"cmdline"`
	Create_time int64
	//	Parent           Process // FIXME: recursive
	Status           string
	Cwd              string
	Username         string
	Uids             []int32
	Gids             []int32
	Terminal         string
	Nice             int32
	Ionice           int32
	Rlimit           []RlimitStat
	Io_counters      Io_countersStat
	Num_ctx_switches int32
	Num_fds          int32
	Num_handles      int32
	Num_Threads      int32
	//	Threads  map[string]string
	Cpu_times CPU_TimesStat
	//	Cpu_percent
	Cpu_affinity   []int32
	Memory_info    Memory_infoStat
	Memory_info_ex map[string]string
	Memory_percent float32
	Memory_maps    []Memory_mapsStat
	//	Children       []Process // FIXME: recursive
	Open_files  []Open_filesStat
	Connections []Net_connectionStat
	Is_running  bool
}

type Open_filesStat struct {
	Path string
	Fd   uint32
}

type Memory_infoStat struct {
	RSS int32 // bytes
	VMS int32 // bytes
}

type Memory_mapsStat struct {
	Path      string
	RSS       int32
	Anonymous int32
	Swap      int32
}

type RlimitStat struct {
	Rresource int32
	Soft      int32
	Hard      int32
}

type Io_countersStat struct {
	Read_count  int32
	Write_count int32
	Read_bytes  int32
	Write_bytes int32
}

func Pids() ([]int32, error) {
	ret := make([]int32, 0)
	procs, err := processes()
	if err != nil {
		return ret, nil
	}

	for _, p := range procs {
		ret = append(ret, p.Pid)
	}

	return ret, nil
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
