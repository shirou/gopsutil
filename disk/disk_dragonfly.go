// SPDX-License-Identifier: BSD-3-Clause
//go:build dragonfly

package disk

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// Offsets into struct devstat from sys/devicestat.h (DEVSTAT_VERSION 4).
const (
	devstatHeader     = 8 // generation counter in front of the array
	devstatNameOffset = 12
	devstatNameLen    = 16
	devstatUnitOffset = 28
	devstatReadBytes  = 32
	devstatWriteBytes = 40
	devstatReads      = 56
	devstatWrites     = 64
	devstatMinSize    = 72

	statvfsBufferSize = 512 // larger than struct statvfs
)

var physicalFilesystems = []string{"ufs", "hammer", "hammer2", "msdos"}

func cString(raw []byte) string {
	if end := bytes.IndexByte(raw, 0); end >= 0 {
		raw = raw[:end]
	}
	return string(raw)
}

func PartitionsWithContext(_ context.Context, all bool) ([]PartitionStat, error) {
	count, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return nil, err
	}
	mounts := make([]unix.Statfs_t, count)
	if count, err = unix.Getfsstat(mounts, unix.MNT_NOWAIT); err != nil {
		return nil, err
	}

	partitions := make([]PartitionStat, 0, count)
	for i := range mounts[:count] {
		fstype := cString(mounts[i].Fstypename[:])
		if !all && !slices.Contains(physicalFilesystems, fstype) {
			continue
		}
		// HAMMER2 reports "da0s1d@ROOT": strip the label to get the device
		device, _, _ := strings.Cut(cString(mounts[i].Mntfromname[:]), "@")
		if !strings.HasPrefix(device, "/") {
			device = "/dev/" + device
		}
		partitions = append(partitions, PartitionStat{
			Device:     device,
			Mountpoint: cString(mounts[i].Mntonname[:]),
			Fstype:     fstype,
		})
	}
	return partitions, nil
}

// statvfs matches what df reports. x/sys/unix has no wrapper for this syscall on
// DragonFly, so it is called directly.
func statvfs(path string) (blockSize, blocks, free, available, files, filesFree uint64, err error) {
	pathPtr, err := unix.BytePtrFromString(path)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	var buf [statvfsBufferSize]byte
	_, _, errno := unix.Syscall(unix.SYS_STATVFS, uintptr(unsafe.Pointer(pathPtr)), uintptr(unsafe.Pointer(&buf[0])), 0)
	if errno != 0 {
		return 0, 0, 0, 0, 0, 0, errno
	}
	field := func(i int) uint64 { return binary.LittleEndian.Uint64(buf[i*8:]) }
	return field(1), field(2), field(3), field(4), field(5), field(6), nil
}

func UsageWithContext(_ context.Context, path string) (*UsageStat, error) {
	blockSize, blocks, free, available, files, filesFree, err := statvfs(path)
	if err != nil {
		return nil, err
	}
	var fs unix.Statfs_t
	if err := unix.Statfs(path, &fs); err != nil {
		return nil, err
	}
	stat := &UsageStat{
		Path:        path,
		Fstype:      cString(fs.Fstypename[:]),
		Total:       blocks * blockSize,
		Free:        available * blockSize,
		Used:        (blocks - free) * blockSize,
		InodesTotal: files,
		InodesFree:  filesFree,
	}
	stat.InodesUsed = stat.InodesTotal - stat.InodesFree
	if usable := stat.Used + stat.Free; usable > 0 {
		stat.UsedPercent = float64(stat.Used) / float64(usable) * 100
	}
	if stat.InodesTotal > 0 {
		stat.InodesUsedPercent = float64(stat.InodesUsed) / float64(stat.InodesTotal) * 100
	}
	return stat, nil
}

func IOCountersWithContext(_ context.Context, names ...string) (map[string]IOCountersStat, error) {
	numDevs, err := unix.SysctlUint32("kern.devstat.numdevs")
	if err != nil {
		return nil, err
	}
	buf, err := unix.SysctlRaw("kern.devstat.all")
	if err != nil {
		return nil, err
	}
	if numDevs == 0 || len(buf) <= devstatHeader {
		return nil, errors.New("kern.devstat.all is empty")
	}
	stride := (len(buf) - devstatHeader) / int(numDevs)
	if stride < devstatMinSize {
		return nil, errors.New("kern.devstat.all has an unknown layout")
	}

	counters := make(map[string]IOCountersStat, numDevs)
	for i := 0; i < int(numDevs); i++ {
		record := buf[devstatHeader+i*stride:]
		unit := int32(binary.LittleEndian.Uint32(record[devstatUnitOffset:]))
		name := fmt.Sprintf("%s%d", cString(record[devstatNameOffset:devstatNameOffset+devstatNameLen]), unit)
		if len(names) > 0 && !slices.Contains(names, name) {
			continue
		}
		counters[name] = IOCountersStat{
			Name:       name,
			ReadBytes:  binary.LittleEndian.Uint64(record[devstatReadBytes:]),
			WriteBytes: binary.LittleEndian.Uint64(record[devstatWriteBytes:]),
			ReadCount:  binary.LittleEndian.Uint64(record[devstatReads:]),
			WriteCount: binary.LittleEndian.Uint64(record[devstatWrites:]),
		}
	}
	return counters, nil
}

func SerialNumberWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}
