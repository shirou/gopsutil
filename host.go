package gopsutil

// A HostInfoStat describes the host status.
// This is not in the psutil but it useful.
type HostInfoStat struct {
	Hostname string `json:"hostname"`
	Uptime   int64  `json:"uptime"`
	Procs    uint64 `json:"procs"`
}
