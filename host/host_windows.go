// +build windows

package host

import (
	"os"
	"time"
	"runtime"
	"strings"

	"github.com/StackExchange/wmi"

	common "github.com/shirou/gopsutil/common"
	process "github.com/shirou/gopsutil/process"
)

var (
	procGetSystemTimeAsFileTime = common.Modkernel32.NewProc("GetSystemTimeAsFileTime")
	osInfo *Win32_OperatingSystem
)

type Win32_OperatingSystem struct {
	Version     string
	Caption     string
	ProductType uint32
	LastBootUpTime time.Time
}

func HostInfo() (*HostInfoStat, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	ret := &HostInfoStat{
		Hostname: hostname,
		OS:       runtime.GOOS,
	}

	platform, family, version, err := GetPlatformInformation()
	if err == nil {
		ret.Platform = platform
		ret.PlatformFamily = family
		ret.PlatformVersion = version
	} else {
	  return ret, err
	}
	
	ret.Uptime, err = BootTime()
	if err != nil {
		return ret, err
	}
	
	procs, err := process.Pids()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}

func GetOSInfo() (Win32_OperatingSystem, error) {
	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	err := wmi.Query(q, &dst)
	if err != nil {
		return Win32_OperatingSystem{}, err
	}
	
	osInfo = &dst[0]
	
	return dst[0], nil
}

func BootTime() (uint64, error) {
	if osInfo == nil {
		_, err := GetOSInfo()
		if err != nil {
			return 0, err
		}
	}
	now := time.Now()
	t := osInfo.LastBootUpTime.Local()
	return uint64(now.Sub(t).Seconds()), nil
}

func GetPlatformInformation() (platform string, family string, version string, err error) {
	if osInfo == nil {
		_, err = GetOSInfo()
		if err != nil {
			return
		}
	}

	// Platform
	platform = strings.Trim(osInfo.Caption, " ")
	
	// PlatformFamily
	switch osInfo.ProductType {
	case 1:
		family = "Desktop OS"
	case 2:
		family = "Server OS (Domain Controller)"
	case 3:
		family = "Server OS"
	}
	
	// Platform Version
	version = osInfo.Version

	return
}

func Users() ([]UserStat, error) {

	var ret []UserStat

	return ret, nil
}
