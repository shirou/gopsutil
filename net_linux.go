// +build linux

package gopsutil

import (
	"strings"
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	filename := "/proc/net/dev"
	lines, err := readLines(filename)
	if err != nil {
		return make([]NetIOCountersStat, 0), err
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
			BytesRecv:   parseUint64(fields[1]),
			Errin:       parseUint64(fields[2]),
			Dropin:      parseUint64(fields[3]),
			BytesSent:   parseUint64(fields[9]),
			PacketsSent: parseUint64(fields[10]),
			Errout:      parseUint64(fields[11]),
			Dropout:     parseUint64(fields[12]),
		}
		ret = append(ret, nic)
	}
	return ret, nil
}
