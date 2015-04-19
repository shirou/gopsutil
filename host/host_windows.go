// +build windows

package host

import (
	"os"
	"time"

	"github.com/StackExchange/wmi"

	common "github.com/shirou/gopsutil/common"
	process "github.com/shirou/gopsutil/process"
)

var (
	procGetSystemTimeAsFileTime = common.Modkernel32.NewProc("GetSystemTimeAsFileTime")
)

type Win32_OperatingSystem struct {
	LastBootUpTime time.Time
}

func HostInfo() (*HostInfoStat, error) {
	ret := &HostInfoStat{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname
	uptime, err := BootTime()
	if err == nil {
		ret.Uptime = uptime
	}

	procs, err := process.Pids()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}

func BootTime() (uint64, error) {
	now := time.Now()

	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		return 0, err
	}
	t := dst[0].LastBootUpTime.Local()
	return uint64(now.Sub(t).Seconds()), nil
}

func Users() ([]UserStat, error) {

	var ret []UserStat

	return ret, nil
}
