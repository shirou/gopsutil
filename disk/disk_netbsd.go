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
    
    flag := uint64(1) // ST_WAIT/MNT_WAIT, see sys/fstypes.h

    // get required buffer size
    r, _, err = unix.Syscall6(
        483, // SYS___getvfsstat90 syscall 
        nil,
        0,
        uintptr(unsafe.Pointer(&flag)), 
    )
    if err != nil {
        return ret, err
    }
    mountedFsCount := uint64(r)

    // calculate the buffer size
    bufSize := sizeOfStatvfs * mountedFsCount
    buf := make([]Statvfs, bufSize)

    // request agian to get desired mount data
    _, _, err = unix.Syscall6(
        483,
        uintptr(unsafe.Pointer(&buf[0])),
        uintptr(unsafe.Pointer(&bufSize)),
        uintptr(unsafe.Pointer(&flag)),
    )
    if err != nil {
        return ret, err
    }

    for _, stat := range buf {
		d := PartitionStat{
			Device:     common.ByteToString([]byte(stat.Mntfromname[:])),
			Mountpoint: common.ByteToString([]byte(stat.Mntonname[:])),
			Fstype:     common.ByteToString([]byte(stat.Fstypename[:])),
			Opts:       opts,
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	ret := make(map[string]IOCountersStat)
	return ret, common.ErrNotImplementedError
}

func UsageWithContext(ctx context.Context, path string) (*UsageStat, error) {
    stat := Statvfs{}
    flag := uint64(1) // ST_WAIT/MNT_WAIT, see sys/fstypes.h

    // request agian to get desired mount data
    ret, _, err = unix.Syscall6(
        485, // SYS___fstatvfs190, see sys/syscall.h
        uintptr(unsafe.Pointer(&path)),
        uintptr(unsafe.Pointer(&stat)),
        uintptr(unsafe.Pointer(&flag)),
    )
    if err != nil {
        return ret, err
    }

    bsize := stat.Bsize
	ret := &UsageStat{
		Path:        path,
		Fstype:      getFsType(stat),
		Total:       (uint64(stat.Blocks) * uint64(bsize)),
		Free:        (uint64(stat.Bavail) * uint64(bsize)),
		InodesTotal: (uint64(stat.Files)),
		InodesFree:  (uint64(stat.Ffree)),
	}

	ret.InodesUsed = (ret.InodesTotal - ret.InodesFree)
	ret.InodesUsedPercent = (float64(ret.InodesUsed) / float64(ret.InodesTotal)) * 100.0
	ret.Used = (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(bsize)
	ret.UsedPercent = (float64(ret.Used) / float64(ret.Total)) * 100.0

	return ret, nil
}

func getFsType(stat Statvfs) string {
	return common.ByteToString(stat.Fstypename[:])
}

func SerialNumberWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}
