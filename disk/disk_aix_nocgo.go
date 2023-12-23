//go:build aix && !cgo
// +build aix,!cgo

package disk

import (
	"context"
	"regexp"
	"strings"

	"github.com/shirou/gopsutil/v3/internal/common"
	"golang.org/x/sys/unix"
)

var startBlank = regexp.MustCompile(`^\s+`)

var ignoreFSType = map[string]bool{"procfs": true}
var FSType = map[int]string{
	0: "jfs2", 1: "namefs", 2: "nfs", 3: "jfs", 5: "cdrom", 6: "proc",
	16: "special-fs", 17: "cache-fs", 18: "nfs3", 19: "automount-fs", 20: "pool-fs", 32: "vxfs",
	33: "veritas-fs", 34: "udfs", 35: "nfs4", 36: "nfs4-pseudo", 37: "smbfs", 38: "mcr-pseudofs",
	39: "ahafs", 40: "sterm-nfs", 41: "asmfs",
}

func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
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
		p := strings.Fields(lines[idx])
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
