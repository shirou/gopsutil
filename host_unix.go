// +build linux freebsd

package gopsutil

import (
	"os"
	"syscall"
)

func HostInfo() (HostInfoStat, error) {
	ret := HostInfoStat{}

	hostname, err := os.Hostname()
	ret.Hostname = hostname
	if err != nil {
		return ret, err
	}

	return ret, nil
}


func Boot_time() (int64, error){
	sysinfo := &syscall.Sysinfo_t{}
	if err := syscall.Sysinfo(sysinfo); err != nil {
		return 0, err
	}
	return sysinfo.Uptime, nil
}
