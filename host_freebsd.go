// +build freebsd

package gopsutil

import (
	"os"
	"strconv"
	"strings"
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

func Boot_time() (int64, error) {
	values, err := do_sysctrl("kern.boottime")
	if err != nil {
		return 0, err
	}
	// ex: { sec = 1392261637, usec = 627534 } Thu Feb 13 12:20:37 2014
	v := strings.Replace(values[2], ",", "", 1)

	boottime, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return boottime, nil
}
