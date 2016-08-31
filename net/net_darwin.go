// +build darwin

package net

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

// example of `netstat -ibdnWI lo0` output on yosemite
// Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll Drop
// lo0   16384 <Link#1>                        869107     0  169411755   869107     0  169411755     0   0
// lo0   16384 ::1/128     ::1                 869107     -  169411755   869107     -  169411755     -   -
// lo0   16384 127           127.0.0.1         869107     -  169411755   869107     -  169411755     -   -
func IOCounters(pernic bool) ([]IOCountersStat, error) {
	const endOfLine = "\n"
	// example of `ifconfig -l` output on yosemite:
	// lo0 gif0 stf0 en0 p2p0 awdl0
	ifconfig, err := exec.LookPath("/sbin/ifconfig")
	if err != nil {
		return nil, err
	}

	netstat, err := exec.LookPath("/usr/sbin/netstat")
	if err != nil {
		return nil, err
	}

	// list all interfaces
	out, err := invoke.Command(ifconfig, "-l")
	if err != nil {
		return nil, err
	}
	interfaces := strings.Fields(strings.TrimRight(string(out), endOfLine))
	ret := make([]IOCountersStat, 0)

	// extract metrics for all interfaces
	for _, interfaceName := range interfaces {
		if out, err = invoke.Command(netstat, "-ibdnWI" + interfaceName); err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), endOfLine)
		if len(lines) <= 1 {
			// invalid output
			continue
		}

		if len(lines[1]) == 0 {
			// interface had been removed since `ifconfig -l` had been executed
			continue
		}

		// only the first output is fine
		values := strings.Fields(lines[1])

		base := 1
		// sometimes Address is ommitted
		if len(values) < 12 {
			base = 0
		}

		parsed := make([]uint64, 0, 7)
		vv := []string{
			values[base+3], // Ipkts == PacketsRecv
			values[base+4], // Ierrs == Errin
			values[base+5], // Ibytes == BytesRecv
			values[base+6], // Opkts == PacketsSent
			values[base+7], // Oerrs == Errout
			values[base+8], // Obytes == BytesSent
		}
		if len(values) == 12 {
			vv = append(vv, values[base+10])
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

		n := IOCountersStat{
			Name:        interfaceName,
			PacketsRecv: parsed[0],
			Errin:       parsed[1],
			BytesRecv:   parsed[2],
			PacketsSent: parsed[3],
			Errout:      parsed[4],
			BytesSent:   parsed[5],
		}
		if len(parsed) == 7 {
			n.Dropout = parsed[6]
		}
		ret = append(ret, n)
	}

	if pernic == false {
		return getIOCountersAll(ret)
	}

	return ret, nil
}

// NetIOCountersByFile is an method which is added just a compatibility for linux.
func IOCountersByFile(pernic bool, filename string) ([]IOCountersStat, error) {
	return IOCounters(pernic)
}

func FilterCounters() ([]FilterStat, error) {
	return nil, errors.New("NetFilterCounters not implemented for darwin")
}

// NetProtoCounters returns network statistics for the entire system
// If protocols is empty then all protocols are returned, otherwise
// just the protocols in the list are returned.
// Not Implemented for Darwin
func ProtoCounters(protocols []string) ([]ProtoCountersStat, error) {
	return nil, errors.New("NetProtoCounters not implemented for darwin")
}
