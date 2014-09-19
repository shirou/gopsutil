// +build linux

package gopsutil

import (
	"strconv"
	"strings"
)

const (
	SectorSize = 512
)

// Get disk partitions.
// should use setmntent(3) but this implement use /etc/mtab file
func DiskPartitions(all bool) ([]DiskPartitionStat, error) {

	filename := "/etc/mtab"
	lines, err := readLines(filename)
	if err != nil {
		return nil, err
	}

	ret := make([]DiskPartitionStat, 0, len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		d := DiskPartitionStat{
			Device:     fields[0],
			Mountpoint: fields[1],
			Fstype:     fields[2],
			Opts:       fields[3],
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	filename := "/proc/diskstats"
	lines, err := readLines(filename)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]DiskIOCountersStat, 0)
	empty := DiskIOCountersStat{}

	for _, line := range lines {
		fields := strings.Fields(line)
		name := fields[2]
		reads, err := strconv.ParseUint((fields[3]), 10, 64)
		if err != nil {
			return ret, err
		}
		rbytes, err := strconv.ParseUint((fields[5]), 10, 64)
		if err != nil {
			return ret, err
		}
		rtime, err := strconv.ParseUint((fields[6]), 10, 64)
		if err != nil {
			return ret, err
		}
		writes, err := strconv.ParseUint((fields[7]), 10, 64)
		if err != nil {
			return ret, err
		}
		wbytes, err := strconv.ParseUint((fields[9]), 10, 64)
		if err != nil {
			return ret, err
		}
		wtime, err := strconv.ParseUint((fields[10]), 10, 64)
		if err != nil {
			return ret, err
		}
		iotime, err := strconv.ParseUint((fields[12]), 10, 64)
		if err != nil {
			return ret, err
		}
		d := DiskIOCountersStat{
			ReadBytes:  rbytes * SectorSize,
			WriteBytes: wbytes * SectorSize,
			ReadCount:  reads,
			WriteCount: writes,
			ReadTime:   rtime,
			WriteTime:  wtime,
			IoTime:     iotime,
		}
		if d == empty {
			continue
		}
		d.Name = name

		ret[name] = d
	}
	return ret, nil
}
