package main

import (
	"os"
	"syscall"
)

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

func (h Host) HostInfo() (HostInfo, error) {
	ret := HostInfo{}
	sysinfo := &syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(sysinfo); err != nil {
		return ret, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}
	ret.Hostname = hostname
	ret.Uptime = sysinfo.Uptime
	ret.Procs = uint64(sysinfo.Procs)

	return ret, nil
}
