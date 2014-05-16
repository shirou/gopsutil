package gopsutil

import (
	"encoding/json"
)

// A HostInfoStat describes the host status.
// This is not in the psutil but it useful.
type HostInfoStat struct {
	Hostname        string `json:"hostname"`
	Uptime          int64  `json:"uptime"`
	Procs           uint64 `json:"procs"` // number of processes
	OS              string `json:"os"` // ex: freebsd, linux
	Platform        string `json:"platform"`        // ex: ubuntu, linuxmint
	PlatformFamily  string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion string `json:"platformVersion"`
}

type UserStat struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  int    `json:"started"`
}

func (h HostInfoStat) String() string {
	s, _ := json.Marshal(h)
	return string(s)
}

func (u UserStat) String() string {
	s, _ := json.Marshal(u)
	return string(s)
}
