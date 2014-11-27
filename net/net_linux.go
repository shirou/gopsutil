// +build linux

package gopsutil

import (
	"strconv"
	"strings"

	common "github.com/shirou/gopsutil/common"
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	filename := "/proc/net/dev"
	lines, err := common.ReadLines(filename)
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

		bytesRecv, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsRecv, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return ret, err
		}
		errIn, err := strconv.ParseUint(fields[3], 10, 64)
		if err != nil {
			return ret, err
		}
		dropIn, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			return ret, err
		}
		bytesSent, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			return ret, err
		}
		packetsSent, err := strconv.ParseUint(fields[10], 10, 64)
		if err != nil {
			return ret, err
		}
		errOut, err := strconv.ParseUint(fields[11], 10, 64)
		if err != nil {
			return ret, err
		}
		dropOut, err := strconv.ParseUint(fields[14], 10, 64)
		if err != nil {
			return ret, err
		}

		nic := NetIOCountersStat{
			Name:        strings.Trim(fields[0], ":"),
			BytesRecv:   bytesRecv,
			PacketsRecv: packetsRecv,
			Errin:       errIn,
			Dropin:      dropIn,
			BytesSent:   bytesSent,
			PacketsSent: packetsSent,
			Errout:      errOut,
			Dropout:     dropOut,
		}
		ret = append(ret, nic)
	}
	return ret, nil
}
