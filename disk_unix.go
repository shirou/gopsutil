// +build freebsd linux darwin

package gopsutil

import "syscall"

func DiskUsage(path string) (*DiskUsageStat, error) {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	bsize := stat.Bsize / 512

	ret := &DiskUsageStat{
		Path:  path,
		Total: (uint64(stat.Blocks) * uint64(bsize)) >> 1,
		Free:  (uint64(stat.Bfree) * uint64(bsize)) >> 1,
		InodesTotal: (uint64(stat.Files)),
		InodesFree: (uint64(stat.Ffree)),
	}

	ret.InodesUsed = (ret.InodesTotal - ret.InodesFree)
	ret.InodesUsedPercent = (float64(ret.InodesUsed) / float64(ret.InodesTotal)) * 100.0
	ret.Used = (ret.Total - ret.Free)
	ret.UsedPercent = (float64(ret.Used) / float64(ret.Total)) * 100.0

	return ret, nil
}
