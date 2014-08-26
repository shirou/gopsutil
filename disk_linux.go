// +build linux

package gopsutil

import (
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
			Device:	    fields[0],
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
		reads := mustParseUint64(fields[3])
		rbytes := mustParseUint64(fields[5])
		rtime := mustParseUint64(fields[6])
		writes := mustParseUint64(fields[7])
		wbytes := mustParseUint64(fields[9])
		wtime := mustParseUint64(fields[10])
		iotime := mustParseUint64(fields[13])
		d := DiskIOCountersStat{
			ReadBytes:  rbytes * SectorSize,
			WriteBytes: wbytes * SectorSize,
			ReadCount:  reads,
			WriteCount: writes,
			ReadTime:   rtime,
			WriteTime:  wtime,
			IoTime:	    iotime,
		}
		if d == empty {
			continue
		}
		d.Name = name

		ret[name] = d
	}
	return ret, nil
}
