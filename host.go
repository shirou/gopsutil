package gopsutil

type HostInfoStat struct {
	Hostname string `json:"hostname"`
	Uptime   int64  `json:"uptime"`
	Procs    uint64 `json:"procs"`
}
