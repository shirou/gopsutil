// +build linux

package load

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/internal/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	filename := common.HostProc("loadavg")
	line, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(line))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	ret := &LoadAvgStat{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}

func Misc() (*MiscStat, error) {
	filename := common.HostProc("stat")
	lines, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ret := &Misc{
		ProcsRunning: pr,
		ProcsBlocked: pb,
		Ctxt:         ctxt,
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		v, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "procs_running":
			ret.ProcessRunning = v
		case "procs_blocked":
			ret.ProcessBlocked = v
		case "ctxt":
			ret.Ctxt = v
		default:
			continue
		}

	}

	return ret, nil
}
