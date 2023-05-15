//go:build darwin
// +build darwin

package disk

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v3/internal/common"
	"golang.org/x/sys/unix"
)

// PartitionsWithContext returns disk partition.
// 'all' argument is ignored, see: https://github.com/giampaolo/psutil/issues/906
func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	var ret []PartitionStat

	count, err := unix.Getfsstat(nil, unix.MNT_WAIT)
	if err != nil {
		return ret, err
	}
	fs := make([]unix.Statfs_t, count)
	count, err = unix.Getfsstat(fs, unix.MNT_WAIT)
	if err != nil {
		return ret, err
	}
	// On 10.14, and possibly other OS versions, the actual count may
	// be less than from the first call. Truncate to the returned count
	// to prevent accessing uninitialized entries.
	// https://github.com/shirou/gopsutil/issues/1390
	fs = fs[:count]
	for _, stat := range fs {
		opts := []string{"rw"}
		if stat.Flags&unix.MNT_RDONLY != 0 {
			opts = []string{"ro"}
		}
		if stat.Flags&unix.MNT_SYNCHRONOUS != 0 {
			opts = append(opts, "sync")
		}
		if stat.Flags&unix.MNT_NOEXEC != 0 {
			opts = append(opts, "noexec")
		}
		if stat.Flags&unix.MNT_NOSUID != 0 {
			opts = append(opts, "nosuid")
		}
		if stat.Flags&unix.MNT_UNION != 0 {
			opts = append(opts, "union")
		}
		if stat.Flags&unix.MNT_ASYNC != 0 {
			opts = append(opts, "async")
		}
		if stat.Flags&unix.MNT_DONTBROWSE != 0 {
			opts = append(opts, "nobrowse")
		}
		if stat.Flags&unix.MNT_AUTOMOUNTED != 0 {
			opts = append(opts, "automounted")
		}
		if stat.Flags&unix.MNT_JOURNALED != 0 {
			opts = append(opts, "journaled")
		}
		if stat.Flags&unix.MNT_MULTILABEL != 0 {
			opts = append(opts, "multilabel")
		}
		if stat.Flags&unix.MNT_NOATIME != 0 {
			opts = append(opts, "noatime")
		}
		if stat.Flags&unix.MNT_NODEV != 0 {
			opts = append(opts, "nodev")
		}
		d := PartitionStat{
			Device:     common.ByteToString(stat.Mntfromname[:]),
			Mountpoint: common.ByteToString(stat.Mntonname[:]),
			Fstype:     common.ByteToString(stat.Fstypename[:]),
			Opts:       opts,
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func getFsType(stat unix.Statfs_t) string {
	return common.ByteToString(stat.Fstypename[:])
}

func SerialNumberWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}

func ModelWithContext(ctx context.Context, name string) (map[string]string, error) {
	out, err := exec.Command("diskutil", "list").Output()
	if err != nil {
		return nil, errors.New("failed to execute 'diskutil list' command: " + err.Error())
	}
	outStr := string(out)
	lines := strings.Split(outStr, "\n")
	diskMap := make(map[string]string)
	for _, line := range lines {
		if strings.HasPrefix(line, "/dev/") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				partitionPath := fields[0]
				if name != "" && !strings.Contains(partitionPath, name) {
					continue
				}
				infoOut, err := exec.Command("diskutil", "info", partitionPath).Output()
				if err != nil {
					return nil, errors.New("failed to execute 'diskutil info' command: " + err.Error())
				}
				infoOutStr := string(infoOut)
				infoLines := strings.Split(infoOutStr, "\n")
				var diskName, model string
				for _, infoLine := range infoLines {
					if strings.HasPrefix(infoLine, "   Device Node:") {
						diskName = strings.TrimSpace(strings.TrimPrefix(infoLine, "   Device Node:"))
					}
					if strings.HasPrefix(infoLine, "   Device / Media Name:") {
						model = strings.TrimSpace(strings.TrimPrefix(infoLine, "   Device / Media Name:"))
					}
				}
				if diskName != "" && model != "" {
					diskMap[diskName] = model
					if name != "" {
						return diskMap, nil
					}
				}
			}
		}
	}
	return diskMap, nil
}
