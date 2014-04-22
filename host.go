package gopsutil

type Host struct{}

type HostInfo struct {
	Hostname string `json:"hostname"`
	Uptime   int64  `json:"uptime"`
	Procs    uint64 `json:"procs"`
}

func NewHost() Host {
	h := Host{}
	return h
}
