// +build windows

package host

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/StackExchange/wmi"

	common "github.com/shirou/gopsutil/common"
	process "github.com/shirou/gopsutil/process"
)

var (
	procGetSystemTimeAsFileTime = common.Modkernel32.NewProc("GetSystemTimeAsFileTime")
	reWindowsName               = regexp.MustCompile("OS Name:\\s+(.+)")
	reWindowsVersion            = regexp.MustCompile("OS Version:\\s+(.+)")
	reWindowsFamily             = regexp.MustCompile("OS Configuration:\\s+(.+)")
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

	platform, family, version, err := GetPlatformInformation()
	if err == nil {
		ret.Platform = platform
		ret.PlatformFamily = family
		ret.PlatformVersion = version
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

func readValue(input string, re *regexp.Regexp, index int) (string, error) {
	items := re.FindStringSubmatch(input)
	if index >= len(items) {
		return "", fmt.Errorf("Tried to read out of bounds index `%d` from items: %#v", index, items)
	}
	return items[index], nil
}

func GetPlatformInformation() (string, string, string, error) {
	platform, family, version := "", "", ""

	out, err := exec.Command("systeminfo").Output()
	if err != nil {
		return platform, family, version, fmt.Errorf("Failed to run systeminfo: %s", err)
	}

	platform, err = readValue(string(out), reWindowsName, 1)
	if err != nil {
		return platform, family, version, fmt.Errorf("Failed reading OS name: %s", err)
	}

	version, err = readValue(string(out), reWindowsVersion, 1)
	if err != nil {
		return platform, family, version, fmt.Errorf("Failed reading OS version: %s", err)
	}

	family, err = readValue(string(out), reWindowsFamily, 1)
	if err != nil {
		return platform, family, version, fmt.Errorf("Failed reading OS family: %s", err)
	}

	return platform, family, version, nil
}
