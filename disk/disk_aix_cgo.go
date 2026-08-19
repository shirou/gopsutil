// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package disk

/*
#include <unistd.h>
*/
import "C"

import (
	"context"

	"github.com/power-devops/perfstat"
)

func IOCountersWithContext(_ context.Context, names ...string) (map[string]IOCountersStat, error) {
	disks, err := perfstat.DiskStat()
	if err != nil {
		return nil, err
	}

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	clkTck := uint64(C.sysconf(C._SC_CLK_TCK))

	ret := make(map[string]IOCountersStat, len(disks))
	for _, d := range disks {
		if len(nameSet) > 0 && !nameSet[d.Name] {
			continue
		}

		ret[d.Name] = IOCountersStat{
			Name:       d.Name,
			ReadCount:  uint64(d.XRate),
			WriteCount: uint64(d.Xfers - d.XRate),
			ReadBytes:  uint64(d.Rblks) * uint64(d.BSize),
			WriteBytes: uint64(d.Wblks) * uint64(d.BSize),
			// perfstat Rserv, Wserv, and WqTime are in nanoseconds;
			// IOCountersStat expects milliseconds.
			ReadTime:       uint64(d.Rserv) / 1_000_000,
			WriteTime:      uint64(d.Wserv) / 1_000_000,
			IoTime:         uint64(d.Time) * 1000 / clkTck, // d.Time is in kernel ticks; convert to ms
			WeightedIO:     uint64(d.WqTime) / 1_000_000,
			IopsInProgress: uint64(d.QDepth),
		}
	}
	return ret, nil
}

func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	f, err := perfstat.FileSystemStat()
	if err != nil {
		return nil, err
	}
	ret := make([]PartitionStat, 0, len(f))

	for _, fs := range f {
		fstyp, exists := FSType[fs.FSType]
		if !exists {
			fstyp = "unknown"
		}
		info := PartitionStat{
			Device:     fs.Device,
			Mountpoint: fs.MountPoint,
			Fstype:     fstyp,
		}
		ret = append(ret, info)
	}

	return ret, err
}
