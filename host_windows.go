// +build windows

package gopsutil

import (
	"github.com/mitchellh/go-ps"
	"os"
	"syscall"
	"unsafe"
)

var (
	procGetSystemTimeAsFileTime = modKernel32.NewProc("GetSystemTimeAsFileTime")
	procGetTickCount            = modKernel32.NewProc("GetTickCount")
)

func HostInfo() (HostInfoStat, error) {
	ret := HostInfoStat{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname
	uptimemsec, _, err := procGetTickCount.Call()
	if uptimemsec == 0 {
		return ret, syscall.GetLastError()
	}

	ret.Uptime = int64(uptimemsec) / 1000

	procs, err := ps.Processes()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}

func Boot_time() (int64, error) {
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

	return int64(pt - uptime), nil
}
func Users() ([]UserStat, error) {

	ret := make([]UserStat, 0)

	return ret, nil
}
