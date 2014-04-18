// +build freebsd linux

package main

import "syscall"

func (d Disk) Disk_usage(path string) (Disk_usage, error) {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return Disk_usage{Path: path}, err
	}

	bsize := stat.Bsize / 512

	ret := Disk_usage{
		Path:      path,
		Total:     (uint64(stat.Blocks) * uint64(bsize)) >> 1,
		Free:      (uint64(stat.Bfree) * uint64(bsize)) >> 1,
		Available: (uint64(stat.Bavail) * uint64(bsize)) >> 1,
	}

	ret.Used = (ret.Total - ret.Free)
	ret.Percent = (float64(ret.Used) / float64(ret.Total)) * 100.0

	return ret, nil
}
