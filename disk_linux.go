// +build linux

package gopsutil

import (
	"strings"
)

const(
	MNT_WAIT   = 1
)

// Get disk partitions.
// should use setmntent(3) but this implement use /etc/mtab file
func Disk_partitions(all bool) ([]Disk_partitionStat, error) {
	ret := make([]Disk_partitionStat, 0)

	filename := "/etc/mtab"
	lines, err := ReadLines(filename)
	if err != nil{
		return ret, err
	}

	for _, line := range lines{
		fields := strings.Fields(line)
		d := Disk_partitionStat{
			Mountpoint: fields[1],
			Fstype:     fields[2],
			Opts:       fields[3],
		}
		ret = append(ret, d)
	}

	return ret, nil
}
