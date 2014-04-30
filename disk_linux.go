// +build linux

package gopsutil

import (
	"strings"
	"unicode"
)

const (
	SECTOR_SIZE = 512
)

// Get disk partitions.
// should use setmntent(3) but this implement use /etc/mtab file
func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat

	filename := "/etc/mtab"
	lines, err := readLines(filename)
	if err != nil {
		return ret, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		d := DiskPartitionStat{
			Mountpoint: fields[1],
			Fstype:     fields[2],
			Opts:       fields[3],
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	ret := make(map[string]DiskIOCountersStat, 0)

	// determine partitions we want to look for
	filename := "/proc/partitions"
	lines, err := readLines(filename)
	if err != nil {
		return ret, err
	}
	var partitions []string

	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		name := []rune(fields[3])

		if unicode.IsDigit(name[len(name)-1]) {
			partitions = append(partitions, fields[3])
		} else {
			// http://code.google.com/p/psutil/issues/detail?id=338
			lenpart := len(partitions)
			if lenpart == 0 || strings.HasPrefix(partitions[lenpart-1], fields[3]) {
				partitions = append(partitions, fields[3])
			}
		}
	}

	filename = "/proc/diskstats"
	lines, err = readLines(filename)
	if err != nil {
		return ret, err
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		name := fields[2]
		reads := parseUint64(fields[3])
		rbytes := parseUint64(fields[5])
		rtime := parseUint64(fields[6])
		writes := parseUint64(fields[7])
		wbytes := parseUint64(fields[9])
		wtime := parseUint64(fields[10])
		if stringContains(partitions, name) {
			d := DiskIOCountersStat{
				Name:       name,
				ReadBytes:  rbytes * SECTOR_SIZE,
				WriteBytes: wbytes * SECTOR_SIZE,
				ReadCount:  reads,
				WriteCount: writes,
				ReadTime:   rtime,
				WriteTime:  wtime,
			}
			ret[name] = d

		}
	}
	return ret, nil
}
