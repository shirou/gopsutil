// +build linux

package gopsutil

import (
	"strings"
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	return fromNetDevFile("/proc/net/dev")
}

func fromNetDevFile(filename string) ([]NetIOCountersStat, error) {
	lines, err := readLines(filename)
	if err != nil {
		return nil, err
	}

	statlen := len(lines) - 1

	ret := make([]NetIOCountersStat, 0, statlen)

	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if fields[0] == "" {
			continue
		}
		nic := NetIOCountersStat{
			Name:        strings.Trim(fields[0], ":"),
			BytesRecv:   mustParseUint64(fields[1]),
			PacketsRecv: mustParseUint64(fields[2]),
			Errin:       mustParseUint64(fields[3]),
			Dropin:      mustParseUint64(fields[4]),
			BytesSent:   mustParseUint64(fields[9]),
			PacketsSent: mustParseUint64(fields[10]),
			Errout:      mustParseUint64(fields[11]),
			Dropout:     mustParseUint64(fields[12]),
		}
		ret = append(ret, nic)
	}
	return ret, nil
}
