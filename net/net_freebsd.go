// +build freebsd

package net

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/common"
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	out, err := exec.Command("/usr/bin/netstat", "-ibdn").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	ret := make([]NetIOCountersStat, 0, len(lines)-1)
	exists := make([]string, 0, len(ret))

	for _, line := range lines {
		values := strings.Fields(line)
		if len(values) < 1 || values[0] == "Name" {
			continue
		}
		if common.StringsHas(exists, values[0]) {
			// skip if already get
			continue
		}
		exists = append(exists, values[0])

		base := 1
		// sometimes Address is ommitted
		if len(values) < 13 {
			base = 0
		}

		parsed := make([]uint64, 0, 8)
		vv := []string{
			values[base+3],  // PacketsRecv
			values[base+4],  // Errin
			values[base+5],  // Dropin
			values[base+6],  // BytesRecvn
			values[base+7],  // PacketSent
			values[base+8],  // Errout
			values[base+9],  // BytesSent
			values[base+11], // Dropout
		}
		for _, target := range vv {
			if target == "-" {
				parsed = append(parsed, 0)
				continue
			}

			t, err := strconv.ParseUint(target, 10, 64)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, t)
		}

		n := NetIOCountersStat{
			Name:        values[0],
			PacketsRecv: parsed[0],
			Errin:       parsed[1],
			Dropin:      parsed[2],
			BytesRecv:   parsed[3],
			PacketsSent: parsed[4],
			Errout:      parsed[5],
			BytesSent:   parsed[6],
			Dropout:     parsed[7],
		}
		ret = append(ret, n)
	}

	if pernic == false {
		return getNetIOCountersAll(ret)
	}

	return ret, nil
}

// Return a list of network connections opened by a process.
func NetConnectionsPid(kind string, pid int32) ([]NetConnectionStat, error) {
	var ret []NetConnectionStat

	args := []string{"-i"}
	switch strings.ToLower(kind) {
	default:
		fallthrough
	case "":
		fallthrough
	case "all":
		fallthrough
	case "inet":
		args = append(args, "tcp", "-i", "udp")
	case "inet4":
		args = append(args, "4")
	case "inet6":
		args = append(args, "6")
	case "tcp":
		args = append(args, "tcp")
	case "tcp4":
		args = append(args, "4tcp")
	case "tcp6":
		args = append(args, "6tcp")
	case "udp":
		args = append(args, "udp")
	case "udp4":
		args = append(args, "6udp")
	case "udp6":
		args = append(args, "6udp")
	case "unix":
		return ret, common.NotImplementedError
	}

	// we can not use -F filter to get all of required information at once.
	r, err := common.CallLsof(invoke, pid, args...)
	if err != nil {
		return nil, err
	}
	for _, rr := range r {
		if strings.HasPrefix(rr, "COMMAND") {
			continue
		}
		n, err := parseNetLine(rr)
		if err != nil {
			// fmt.Println(err) // TODO: should debug print?
			continue
		}

		ret = append(ret, n)
	}

	return ret, nil
}
