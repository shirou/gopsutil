// +build freebsd

package gopsutil

import (
	"os/exec"
	"strconv"
	"strings"
)

func LoadAvg() (LoadAvgStat, error) {
	out, err := exec.Command("/sbin/sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		return LoadAvgStat{}, err
	}
	v := strings.Replace(string(out), "{ ", "", 1)
	v = strings.Replace(string(v), " }", "", 1)
	values := strings.Fields(string(v))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return LoadAvgStat{}, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return LoadAvgStat{}, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return LoadAvgStat{}, err
	}

	ret := LoadAvgStat{
		Load1:  float64(load1),
		Load5:  float64(load5),
		Load15: float64(load15),
	}

	return ret, nil
}
