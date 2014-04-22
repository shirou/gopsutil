// +build linux freebsd

package gopsutil

import (
	"os"
	"syscall"
)

func (h Host) HostInfo() (HostInfo, error) {
	ret := HostInfo{}
	hostname, err := os.Hostname()
	ret.Hostname = hostname
	ret.Uptime = sysinfo.Uptime

	return ret, nil
}
