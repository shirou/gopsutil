// +build windows

package main

import (
	"github.com/mitchellh/go-ps"
	"os"
	"syscall"
)

func (h Host) HostInfo() (HostInfo, error) {
	ret := HostInfo{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname

	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return ret, err
	}
	defer syscall.FreeLibrary(kernel32)
	GetTickCount, _ := syscall.GetProcAddress(kernel32, "GetTickCount")

	uptimemsec, _, err := syscall.Syscall(uintptr(GetTickCount), 0, 0, 0, 0)

	ret.Uptime = int64(uptimemsec) / 1000

	procs, err := ps.Processes()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}
