// +build linux

package gopsutil

import (
	"strings"
)

func Net_io_counters() ([]Net_io_countersStat, error) {
	filename := "/proc/net/dev"
	lines, err := ReadLines(filename)
	if err != nil {
		return make([]Net_io_countersStat, 0), err
	}

	statlen := len(lines) - 1

	ret := make([]Net_io_countersStat, 0, statlen)

	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if fields[0] == "" {
			continue
		}
		nic := Net_io_countersStat{
			Name:         strings.Trim(fields[0], ":"),
			Bytes_recv:   parseUint64(fields[1]),
			Errin:        parseUint64(fields[2]),
			Dropin:       parseUint64(fields[3]),
			Bytes_sent:   parseUint64(fields[9]),
			Packets_sent: parseUint64(fields[10]),
			Errout:       parseUint64(fields[11]),
			Dropout:      parseUint64(fields[12]),
		}
		ret = append(ret, nic)
	}
	return ret, nil
}
