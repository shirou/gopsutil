// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package disk

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var startBlank = regexp.MustCompile(`^\s+`)

var ignoreFSType = map[string]bool{"procfs": true}

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
