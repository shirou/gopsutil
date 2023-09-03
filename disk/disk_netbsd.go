//go:build openbsd
// +build openbsd

package disk

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/shirou/gopsutil/v3/internal/common"
	"golang.org/x/sys/unix"
)

func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	var ret []PartitionStat

    // get required buffer size
    r, _, err = unix.Syscall6(
        483, // SYS___getvfsstat90 syscall 
        nil,
        0,
        1, // ST_WAIT/MNT_WAIT, see sys/fstypes.h
    )
    if err != nil {
        return ret, err
    }
    mountedFsCount := uint64(r)

		d := PartitionStat{
			Device:     common.ByteToString(stat.F_mntfromname[:]),
			Mountpoint: common.ByteToString(stat.F_mntonname[:]),
			Fstype:     common.ByteToString(stat.F_fstypename[:]),
			Opts:       opts,
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	ret := make(map[string]IOCountersStat)

	r, err := unix.SysctlRaw("hw.diskstats")
	if err != nil {
		return nil, err
	}
	buf := []byte(r)
	length := len(buf)

	count := int(uint64(length) / uint64(sizeOfDiskstats))

	// parse buf to Diskstats
	for i := 0; i < count; i++ {
		b := buf[i*sizeOfDiskstats : i*sizeOfDiskstats+sizeOfDiskstats]
		d, err := parseDiskstats(b)
		if err != nil {
			continue
		}
		name := common.IntToString(d.Name[:])

		if len(names) > 0 && !common.StringsHas(names, name) {
			continue
		}

		ds := IOCountersStat{
			ReadCount:  d.Rxfer,
			WriteCount: d.Wxfer,
			ReadBytes:  d.Rbytes,
			WriteBytes: d.Wbytes,
			Name:       name,
		}
		ret[name] = ds
	}

	return ret, nil
}

// BT2LD(time)     ((long double)(time).sec + (time).frac * BINTIME_SCALE)

func parseDiskstats(buf []byte) (Diskstats, error) {
	var ds Diskstats
	br := bytes.NewReader(buf)
	//	err := binary.Read(br, binary.LittleEndian, &ds)
	err := common.Read(br, binary.LittleEndian, &ds)
	if err != nil {
		return ds, err
	}

	return ds, nil
}

func UsageWithContext(ctx context.Context, path string) (*UsageStat, error) {
	stat := unix.Statfs_t{}
	err := unix.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	bsize := stat.F_bsize

	ret := &UsageStat{
		Path:        path,
		Fstype:      getFsType(stat),
		Total:       (uint64(stat.F_blocks) * uint64(bsize)),
		Free:        (uint64(stat.F_bavail) * uint64(bsize)),
		InodesTotal: (uint64(stat.F_files)),
		InodesFree:  (uint64(stat.F_ffree)),
	}

	ret.InodesUsed = (ret.InodesTotal - ret.InodesFree)
	ret.InodesUsedPercent = (float64(ret.InodesUsed) / float64(ret.InodesTotal)) * 100.0
	ret.Used = (uint64(stat.F_blocks) - uint64(stat.F_bfree)) * uint64(bsize)
	ret.UsedPercent = (float64(ret.Used) / float64(ret.Total)) * 100.0

	return ret, nil
}

func getFsType(stat unix.Statfs_t) string {
	return common.ByteToString(stat.F_fstypename[:])
}

func SerialNumberWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}
