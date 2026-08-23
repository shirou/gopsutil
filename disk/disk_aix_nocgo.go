// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package disk

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var startBlank = regexp.MustCompile(`^\s+`)

var (
	ignoreFSType = map[string]bool{"procfs": true}
	FSType       = map[int]string{
		0: "jfs2", 1: "namefs", 2: "nfs", 3: "jfs", 5: "cdrom", 6: "proc",
		16: "special-fs", 17: "cache-fs", 18: "nfs3", 19: "automount-fs", 20: "pool-fs", 32: "vxfs",
		33: "veritas-fs", 34: "udfs", 35: "nfs4", 36: "nfs4-pseudo", 37: "smbfs", 38: "mcr-pseudofs",
		39: "ahafs", 40: "sterm-nfs", 41: "asmfs",
	}
)

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	out, err := invoke.CommandWithContext(ctx, "iostat", "-d")
	if err != nil {
		return nil, err
	}

	ret := make(map[string]IOCountersStat)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		// Skip the header line
		if fields[0] == "Disks:" {
			continue
		}

		name := fields[0]
		if len(names) > 0 && !common.StringsHas(names, name) {
			continue
		}

		kbRead, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			continue
		}
		kbWritten, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			continue
		}

		ret[name] = IOCountersStat{
			Name:       name,
			ReadBytes:  kbRead * 1024,
			WriteBytes: kbWritten * 1024,
		}
	}

	return ret, nil
}

func PartitionsWithContext(ctx context.Context, _ bool) ([]PartitionStat, error) {
	var ret []PartitionStat

	out, err := invoke.CommandWithContext(ctx, "mount")
	if err != nil {
		return nil, err
	}

	// parse head lines for column names
	colidx := make(map[string]int)
	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return nil, common.ErrNotImplementedError
	}

	idx := 0
	start := 0
	finished := false
	for pos, ch := range lines[1] {
		if ch == ' ' && !finished {
			name := strings.TrimSpace(lines[0][start:pos])
			colidx[name] = idx
			finished = true
		} else if ch == '-' && finished {
			idx++
			start = pos
			finished = false
		}
	}
	name := strings.TrimSpace(lines[0][start:len(lines[1])])
	colidx[name] = idx

	for idx := 2; idx < len(lines); idx++ {
		line := lines[idx]
		if startBlank.MatchString(line) {
			line = "localhost" + line
		}
		p := strings.Fields(line)
		if len(p) < 5 || ignoreFSType[p[colidx["vfs"]]] {
			continue
		}
		d := PartitionStat{
			Device:     p[colidx["mounted"]],
			Mountpoint: p[colidx["mounted over"]],
			Fstype:     p[colidx["vfs"]],
			Opts:       strings.Split(p[colidx["options"]], ","),
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func getFsType(stat unix.Statfs_t) string {
	return FSType[int(stat.Vfstype)]
}

func getUsage(ctx context.Context, path string) (*UsageStat, error) {
	out, err := invoke.CommandWithContext(ctx, "df", "-v")
	if err != nil {
		return nil, err
	}

	blocksize := uint64(512)
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return nil, common.ErrNotImplementedError
	}

	hf := strings.Fields(strings.ReplaceAll(lines[0], "Mounted on", "Path")) // headers

	// Find the Path column index once before the loop.
	pathIdx := -1
	for i, header := range hf {
		if header == "Path" {
			pathIdx = i
			break
		}
	}

	for line := 1; line < len(lines); line++ {
		fs := strings.Fields(lines[line]) // values
		if len(fs) < len(hf) {
			continue
		}

		if pathIdx < 0 || fs[pathIdx] != path {
			continue
		}

		ret := &UsageStat{}
		for i, header := range hf {
			if i >= len(fs) {
				break
			}

			switch header {
			case `Path`:
				ret.Path = fs[i]
			case `512-blocks`:
				if fs[i] == "-" {
					continue
				}
				total, err := strconv.ParseUint(fs[i], 10, 64)
				if err != nil {
					return nil, err
				}
				ret.Total = total * blocksize
			case `Used`:
				if fs[i] == "-" {
					continue
				}
				used, err := strconv.ParseUint(fs[i], 10, 64)
				if err != nil {
					return nil, err
				}
				ret.Used = used * blocksize
			case `Free`:
				if fs[i] == "-" {
					continue
				}
				free, err := strconv.ParseUint(fs[i], 10, 64)
				if err != nil {
					return nil, err
				}
				ret.Free = free * blocksize
			case `%Used`:
				if fs[i] == "-" {
					continue
				}
				val, err := strconv.ParseInt(strings.ReplaceAll(fs[i], "%", ""), 10, 32)
				if err != nil {
					return nil, err
				}
				ret.UsedPercent = float64(val)
			case `Ifree`:
				if fs[i] == "-" {
					continue
				}
				ret.InodesFree, err = strconv.ParseUint(fs[i], 10, 64)
				if err != nil {
					return nil, err
				}
			case `Iused`:
				if fs[i] == "-" {
					continue
				}
				ret.InodesUsed, err = strconv.ParseUint(fs[i], 10, 64)
				if err != nil {
					return nil, err
				}
			case `%Iused`:
				if fs[i] == "-" {
					continue
				}
				val, err := strconv.ParseInt(strings.ReplaceAll(fs[i], "%", ""), 10, 32)
				if err != nil {
					return nil, err
				}
				ret.InodesUsedPercent = float64(val)
			}
		}

		ret.InodesTotal = ret.InodesUsed + ret.InodesFree

		var statfs unix.Statfs_t
		if err := unix.Statfs(path, &statfs); err == nil {
			ret.Fstype = getFsType(statfs)
		}

		return ret, nil
	}

	return nil, fmt.Errorf("mountpoint %s not found", path)
}

func GetMountFSTypeWithContext(ctx context.Context, mp string) (string, error) {
	out, err := invoke.CommandWithContext(ctx, "mount")
	if err != nil {
		return "", err
	}

	// Kind of inefficient, but it works
	lines := strings.Split(string(out), "\n")
	for line := 1; line < len(lines); line++ {
		fields := strings.Fields(lines[line])
		if strings.TrimSpace(fields[0]) == mp {
			return fields[2], nil
		}
	}

	return "", nil
}
