// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package disk

/*
#include <unistd.h>
*/
import "C"

import (
	"context"
	"fmt"

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

var FSType map[int]string

func init() {
	FSType = map[int]string{
		0: "jfs2", 1: "namefs", 2: "nfs", 3: "jfs", 5: "cdrom", 6: "proc",
		16: "special-fs", 17: "cache-fs", 18: "nfs3", 19: "automount-fs", 20: "pool-fs", 32: "vxfs",
		33: "veritas-fs", 34: "udfs", 35: "nfs4", 36: "nfs4-pseudo", 37: "smbfs", 38: "mcr-pseudofs",
		39: "ahafs", 40: "sterm-nfs", 41: "asmfs",
	}
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

func getUsage(_ context.Context, path string) (*UsageStat, error) {
	f, err := perfstat.FileSystemStat()
	if err != nil {
		return nil, err
	}

	blocksize := uint64(512)
	for _, fs := range f {
		if path == fs.MountPoint {
			fstyp, exists := FSType[fs.FSType]
			if !exists {
				fstyp = "unknown"
			}
			info := UsageStat{
				Path:        path,
				Fstype:      fstyp,
				Total:       uint64(fs.TotalBlocks) * blocksize,
				Free:        uint64(fs.FreeBlocks) * blocksize,
				Used:        uint64(fs.TotalBlocks-fs.FreeBlocks) * blocksize,
				InodesTotal: uint64(fs.TotalInodes),
				InodesFree:  uint64(fs.FreeInodes),
				InodesUsed:  uint64(fs.TotalInodes - fs.FreeInodes),
			}
			info.UsedPercent = (float64(info.Used) / float64(info.Total)) * 100.0
			info.InodesUsedPercent = (float64(info.InodesUsed) / float64(info.InodesTotal)) * 100.0
			return &info, nil
		}
	}
	return nil, fmt.Errorf("mountpoint %s not found", path)
}
