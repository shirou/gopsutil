// +build windows

package gopsutil

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	procGetSystemTimeAsFileTime = modkernel32.NewProc("GetSystemTimeAsFileTime")
	procGetTickCount            = modkernel32.NewProc("GetTickCount")
)

func HostInfo() (*HostInfoStat, error) {
	ret := &HostInfoStat{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname
	uptimemsec, _, err := procGetTickCount.Call()
	if uptimemsec == 0 {
		return ret, syscall.GetLastError()
	}

	ret.Uptime = uint64(uptimemsec) / 1000

	procs, err := Pids()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}

func BootTime() (uint64, error) {
	var lpSystemTimeAsFileTime FILETIME

	r, _, _ := procGetSystemTimeAsFileTime.Call(uintptr(unsafe.Pointer(&lpSystemTimeAsFileTime)))
	if r == 0 {
		return 0, syscall.GetLastError()
	}

	// TODO: This calc is wrong.
	ll := (uint32(lpSystemTimeAsFileTime.DwHighDateTime))<<32 + lpSystemTimeAsFileTime.DwLowDateTime
	pt := (uint64(ll) - 116444736000000000) / 10000000

	u, _, _ := procGetTickCount.Call()
	if u == 0 {
		return 0, syscall.GetLastError()
	}
	uptime := uint64(u) / 1000

	return uint64(pt - uptime), nil
}
func Users() ([]UserStat, error) {

	var ret []UserStat

	return ret, nil
}
