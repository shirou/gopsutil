// +build darwin

package net

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/common"
)

var constMap = map[string]int{
	"TCP":  syscall.SOCK_STREAM,
	"UDP":  syscall.SOCK_DGRAM,
	"IPv4": syscall.AF_INET,
	"IPv6": syscall.AF_INET6,
}

// example of netstat -idbn output on yosemite
// Name  Mtu   Network       Address            Ipkts Ierrs     Ibytes    Opkts Oerrs     Obytes  Coll Drop
// lo0   16384 <Link#1>                        869107     0  169411755   869107     0  169411755     0   0
// lo0   16384 ::1/128     ::1                 869107     -  169411755   869107     -  169411755     -   -
// lo0   16384 127           127.0.0.1         869107     -  169411755   869107     -  169411755     -   -
func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	out, err := exec.Command("/usr/sbin/netstat", "-ibdn").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	ret := make([]NetIOCountersStat, 0, len(lines)-1)
	exists := make([]string, 0, len(ret))

	for _, line := range lines {
		values := strings.Fields(line)
		if len(values) < 1 || values[0] == "Name" {
			// skip first line
			continue
		}
		if common.StringsHas(exists, values[0]) {
			// skip if already get
			continue
		}
		exists = append(exists, values[0])

		base := 1
		// sometimes Address is ommitted
		if len(values) < 11 {
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

		n := NetIOCountersStat{
			Name:        values[0],
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
		return getNetIOCountersAll(ret)
	}

	return ret, nil
}

// Return a list of network connections opened by a process
func NetConnections(kind string) ([]NetConnectionStat, error) {
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
		args = append(args, "tcp")
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
	r, err := common.CallLsof(invoke, 0, args...)
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

func parseNetLine(line string) (NetConnectionStat, error) {
	f := strings.Fields(line)
	if len(f) < 9 {
		return NetConnectionStat{}, fmt.Errorf("wrong line,%s", line)
	}

	pid, err := strconv.Atoi(f[1])
	if err != nil {
		return NetConnectionStat{}, err
	}
	fd, err := strconv.Atoi(strings.Trim(f[3], "u"))
	if err != nil {
		return NetConnectionStat{}, fmt.Errorf("unknown fd, %s", f[3])
	}
	netFamily, ok := constMap[f[4]]
	if !ok {
		return NetConnectionStat{}, fmt.Errorf("unknown family, %s", f[4])
	}
	netType, ok := constMap[f[7]]
	if !ok {
		return NetConnectionStat{}, fmt.Errorf("unknown type, %s", f[7])
	}

	laddr, raddr, err := parseNetAddr(f[8])
	if err != nil {
		return NetConnectionStat{}, fmt.Errorf("failed to parse netaddr, %s", f[8])
	}

	n := NetConnectionStat{
		Fd:     uint32(fd),
		Family: uint32(netFamily),
		Type:   uint32(netType),
		Laddr:  laddr,
		Raddr:  raddr,
		Pid:    int32(pid),
	}
	if len(f) == 10 {
		n.Status = strings.Trim(f[9], "()")
	}

	return n, nil
}

func parseNetAddr(line string) (laddr Addr, raddr Addr, err error) {
	parse := func(l string) (Addr, error) {
		host, port, err := net.SplitHostPort(l)
		if err != nil {
			return Addr{}, fmt.Errorf("wrong addr, %s", l)
		}
		lport, err := strconv.Atoi(port)
		if err != nil {
			return Addr{}, err
		}
		return Addr{IP: host, Port: uint32(lport)}, nil
	}

	addrs := strings.Split(line, "->")
	if len(addrs) == 0 {
		return laddr, raddr, fmt.Errorf("wrong netaddr, %s", line)
	}
	laddr, err = parse(addrs[0])
	if len(addrs) == 2 { // remote addr exists
		raddr, err = parse(addrs[1])
		if err != nil {
			return laddr, raddr, err
		}
	}

	return laddr, raddr, err
}
