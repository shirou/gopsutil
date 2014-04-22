// +build linux

package gopsutil

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func (l Load) LoadAvg() (LoadAvg, error) {
	filename := "/proc/loadavg"
	line, err := ioutil.ReadFile(filename)
	if err != nil {
		return LoadAvg{}, err
	}

	values := strings.Fields(string(line))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return LoadAvg{}, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return LoadAvg{}, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return LoadAvg{}, err
	}

	ret := LoadAvg{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}
