// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd || linux || darwin || aix

package disk

import (
	"context"
	"strings"

	"golang.org/x/sys/unix"
)

func UsageWithContext(_ context.Context, path string) (*UsageStat, error) {
	stat := unix.Statfs_t{}
	err := unix.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	bsize := stat.Bsize

	ret := &UsageStat{
		Path:        unescapeFstab(path),
		Fstype:      getFsType(stat),
		Total:       (uint64(stat.Blocks) * uint64(bsize)),
		Free:        (uint64(stat.Bavail) * uint64(bsize)),
		InodesTotal: uint64(stat.Files),
		InodesFree:  uint64(stat.Ffree),
	}

	ret.Used = (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(bsize)

	if (ret.Used + ret.Free) == 0 {
		ret.UsedPercent = 0
	} else {
		// We don't use ret.Total to calculate percent.
		// see https://github.com/shirou/gopsutil/issues/562
		ret.UsedPercent = (float64(ret.Used) / float64(ret.Used+ret.Free)) * 100.0
	}

	// if could not get InodesTotal, return empty
	if ret.InodesTotal < ret.InodesFree {
		return ret, nil
	}

	ret.InodesUsed = (ret.InodesTotal - ret.InodesFree)

	if ret.InodesTotal == 0 {
		ret.InodesUsedPercent = 0
	} else {
		ret.InodesUsedPercent = (float64(ret.InodesUsed) / float64(ret.InodesTotal)) * 100.0
	}

	return ret, nil
}

// unescapeFstab decodes the \ooo octal escapes used in fstab, /proc/mounts and
// /proc/self/mountinfo fields (e.g. space as \040, backslash as \134), see
// issue#555 and getmntent(3). Any other byte, including '"' and bytes that are
// not valid UTF-8, is kept as is.
func unescapeFstab(path string) string {
	if !strings.Contains(path, `\`) {
		return path
	}
	buf := make([]byte, 0, len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' && i+3 < len(path) && isOctalEscape(path[i+1:i+4]) {
			buf = append(buf, (path[i+1]-'0')<<6|(path[i+2]-'0')<<3|(path[i+3]-'0'))
			i += 3
			continue
		}
		buf = append(buf, path[i])
	}
	return string(buf)
}

func isOctalEscape(s string) bool {
	return s[0] >= '0' && s[0] <= '3' &&
		s[1] >= '0' && s[1] <= '7' &&
		s[2] >= '0' && s[2] <= '7'
}
