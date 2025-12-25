// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package disk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// PartitionsWithContext returns disk partition.
// 'all' argument is ignored, see: https://github.com/giampaolo/psutil/issues/906
func PartitionsWithContext(_ context.Context, _ bool) ([]PartitionStat, error) {
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
	for i := range fs {
		stat := &fs[i]
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
		if stat.Flags&unix.MNT_LOCAL != 0 {
			opts = append(opts, "local")
		}
		if stat.Flags&unix.MNT_CPROTECT != 0 {
			opts = append(opts, "protect")
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

type spnvmeDataTypeItem struct {
	Name              string `json:"_name"`
	BsdName           string `json:"bsd_name"`
	DetachableDrive   string `json:"detachable_drive"`
	DeviceModel       string `json:"device_model"`
	DeviceRevision    string `json:"device_revision"`
	DeviceSerial      string `json:"device_serial"`
	PartitionMapType  string `json:"partition_map_type"`
	RemovableMedia    string `json:"removable_media"`
	Size              string `json:"size"`
	SizeInBytes       int64  `json:"size_in_bytes"`
	SmartStatus       string `json:"smart_status"`
	SpnvmeTrimSupport string `json:"spnvme_trim_support"`
	Volumes           []struct {
		Name        string `json:"_name"`
		BsdName     string `json:"bsd_name"`
		Iocontent   string `json:"iocontent"`
		Size        string `json:"size"`
		SizeInBytes int    `json:"size_in_bytes"`
	} `json:"volumes"`
}

type spnvmeDataWrapper struct {
	SPNVMeDataType []struct {
		Items []spnvmeDataTypeItem `json:"_items"`
	} `json:"SPNVMeDataType"`
}

func SerialNumberWithContext(ctx context.Context, _ string) (string, error) {
	output, err := invoke.CommandWithContext(ctx, "system_profiler", "SPNVMeDataType", "-json")
	if err != nil {
		return "", err
	}

	var data spnvmeDataWrapper
	if err := json.Unmarshal(output, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Extract all serial numbers into a single string
	var serialNumbers []string
	for i := range data.SPNVMeDataType {
		spnvmeData := &data.SPNVMeDataType[i]
		for j := range spnvmeData.Items {
			item := &spnvmeData.Items[j]
			serialNumbers = append(serialNumbers, item.DeviceSerial)
		}
	}

	if len(serialNumbers) == 0 {
		return "", errors.New("no serial numbers found")
	}

	return strings.Join(serialNumbers, ", "), nil
}

func LabelWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}

func IOCountersWithContext(_ context.Context, names ...string) (map[string]IOCountersStat, error) {
	iokit, err := common.NewIOKitLib()
	if err != nil {
		return nil, err
	}
	defer iokit.Close()

	corefoundation, err := common.NewCoreFoundationLib()
	if err != nil {
		return nil, err
	}
	defer corefoundation.Close()

	match := iokit.IOServiceMatching("IOMedia")

	key := corefoundation.CFStringCreateWithCString(common.KCFAllocatorDefault, common.KIOMediaWholeKey, common.KCFStringEncodingUTF8)
	defer corefoundation.CFRelease(uintptr(key))

	kCFBooleanTruePtr, _ := corefoundation.Dlsym("kCFBooleanTrue")
	kCFBooleanTrue := **(**uintptr)(unsafe.Pointer(&kCFBooleanTruePtr))
	corefoundation.CFDictionaryAddValue(uintptr(match), uintptr(key), kCFBooleanTrue)

	var drives uint32
	if status := iokit.IOServiceGetMatchingServices(common.KIOMainPortDefault, uintptr(match), &drives); status != common.KERN_SUCCESS {
		return nil, fmt.Errorf("IOServiceGetMatchingServices error=%d", status)
	}
	defer iokit.IOObjectRelease(drives)

	ic := &ioCounters{
		iokit:          iokit,
		corefoundation: corefoundation,
	}

	stats := make([]IOCountersStat, 0, 16)
	for {
		d := iokit.IOIteratorNext(drives)
		if d <= 0 {
			break
		}

		stat, err := ic.getDriveStat(d)
		if err != nil {
			return nil, err
		}

		if stat != nil {
			stats = append(stats, *stat)
		}

		iokit.IOObjectRelease(d)
	}

	ret := make(map[string]IOCountersStat, 0)
	for i := 0; i < len(stats); i++ {
		if len(names) > 0 && !common.StringsHas(names, stats[i].Name) {
			continue
		}

		stats[i].ReadTime = stats[i].ReadTime / 1000 / 1000 // note: read/write time are in ns, but we want ms.
		stats[i].WriteTime = stats[i].WriteTime / 1000 / 1000
		stats[i].IoTime = stats[i].ReadTime + stats[i].WriteTime

		ret[stats[i].Name] = stats[i]
	}

	return ret, nil
}

const (
	kIOBSDNameKey = "BSD Name"
	// kIOMediaSizeKey               = "Size"
	// kIOMediaPreferredBlockSizeKey = "Preferred Block Size"

	kIOBlockStorageDriverStatisticsKey               = "Statistics"
	kIOBlockStorageDriverStatisticsBytesReadKey      = "Bytes (Read)"
	kIOBlockStorageDriverStatisticsBytesWrittenKey   = "Bytes (Write)"
	kIOBlockStorageDriverStatisticsReadsKey          = "Operations (Read)"
	kIOBlockStorageDriverStatisticsWritesKey         = "Operations (Write)"
	kIOBlockStorageDriverStatisticsTotalReadTimeKey  = "Total Time (Read)"
	kIOBlockStorageDriverStatisticsTotalWriteTimeKey = "Total Time (Write)"
)

type ioCounters struct {
	iokit          *common.IOKitLib
	corefoundation *common.CoreFoundationLib
}

func (i *ioCounters) getDriveStat(d uint32) (*IOCountersStat, error) {
	var parent uint32
	if status := i.iokit.IORegistryEntryGetParentEntry(d, common.KIOServicePlane, &parent); status != common.KERN_SUCCESS {
		return nil, fmt.Errorf("IORegistryEntryGetParentEntry error=%d", status)
	}
	defer i.iokit.IOObjectRelease(parent)

	if !i.iokit.IOObjectConformsTo(parent, "IOBlockStorageDriver") {
		// return nil, fmt.Errorf("ERROR: the object is not of the IOBlockStorageDriver class")
		return nil, nil
	}

	var props unsafe.Pointer
	if status := i.iokit.IORegistryEntryCreateCFProperties(d, unsafe.Pointer(&props), common.KCFAllocatorDefault, common.KNilOptions); status != common.KERN_SUCCESS {
		return nil, fmt.Errorf("IORegistryEntryCreateCFProperties error=%d", status)
	}
	defer i.corefoundation.CFRelease(uintptr(props))

	key := i.cfStr(kIOBSDNameKey)
	defer i.corefoundation.CFRelease(uintptr(key))
	name := i.corefoundation.CFDictionaryGetValue(uintptr(props), uintptr(key))

	buf := common.NewCStr(i.corefoundation.CFStringGetLength(uintptr(name)))
	i.corefoundation.CFStringGetCString(uintptr(name), buf, buf.Length(), common.KCFStringEncodingUTF8)

	stat, err := i.fillStat(parent)
	if err != nil {
		return nil, err
	}

	if stat != nil {
		stat.Name = buf.GoString()
		return stat, nil
	}
	return nil, nil
}

func (i *ioCounters) fillStat(d uint32) (*IOCountersStat, error) {
	var props unsafe.Pointer
	status := i.iokit.IORegistryEntryCreateCFProperties(d, unsafe.Pointer(&props), common.KCFAllocatorDefault, common.KNilOptions)
	if status != common.KERN_SUCCESS {
		return nil, fmt.Errorf("IORegistryEntryCreateCFProperties error=%d", status)
	}
	if props == nil {
		return nil, nil
	}
	defer i.corefoundation.CFRelease(uintptr(props))

	key := i.cfStr(kIOBlockStorageDriverStatisticsKey)
	defer i.corefoundation.CFRelease(uintptr(key))

	v := i.corefoundation.CFDictionaryGetValue(uintptr(props), uintptr(key))
	if v == nil {
		return nil, errors.New("CFDictionaryGetValue failed")
	}

	var stat IOCountersStat
	statstab := map[string]uintptr{
		kIOBlockStorageDriverStatisticsBytesReadKey:      unsafe.Offsetof(stat.ReadBytes),
		kIOBlockStorageDriverStatisticsBytesWrittenKey:   unsafe.Offsetof(stat.WriteBytes),
		kIOBlockStorageDriverStatisticsReadsKey:          unsafe.Offsetof(stat.ReadCount),
		kIOBlockStorageDriverStatisticsWritesKey:         unsafe.Offsetof(stat.WriteCount),
		kIOBlockStorageDriverStatisticsTotalReadTimeKey:  unsafe.Offsetof(stat.ReadTime),
		kIOBlockStorageDriverStatisticsTotalWriteTimeKey: unsafe.Offsetof(stat.WriteTime),
	}

	for key, off := range statstab {
		s := i.cfStr(key)
		if num := i.corefoundation.CFDictionaryGetValue(uintptr(v), uintptr(s)); num != nil {
			i.corefoundation.CFNumberGetValue(uintptr(num), common.KCFNumberSInt64Type, uintptr(unsafe.Add(unsafe.Pointer(&stat), off)))
		}
		i.corefoundation.CFRelease(uintptr(s))
	}

	return &stat, nil
}

func (i *ioCounters) cfStr(str string) unsafe.Pointer {
	return i.corefoundation.CFStringCreateWithCString(common.KCFAllocatorDefault, str, common.KCFStringEncodingUTF8)
}
