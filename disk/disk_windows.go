// +build windows

package disk

import (
	"bytes"
	"context"
	"errors"
	"unsafe"

	"github.com/shirou/gopsutil/internal/common"
	"golang.org/x/sys/windows"
)

var (
	procGetDiskFreeSpaceExW = common.Modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetDriveType        = common.Modkernel32.NewProc("GetDriveTypeW")
)

var (
	FileFileCompression = uint32(16)     // 0x00000010
	FileReadOnlyVolume  = uint32(524288) // 0x00080000
)

type Win32_PerfFormattedData struct {
	Name                    string
	AvgDiskBytesPerRead     uint64
	AvgDiskBytesPerWrite    uint64
	AvgDiskReadQueueLength  uint64
	AvgDiskWriteQueueLength uint64
	AvgDisksecPerRead       uint64
	AvgDisksecPerWrite      uint64
}

const WaitMSec = 500

func Usage(path string) (*UsageStat, error) {
	return UsageWithContext(context.Background(), path)
}

func UsageWithContext(ctx context.Context, path string) (*UsageStat, error) {
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	diskret, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	if diskret == 0 {
		return nil, err
	}
	ret := &UsageStat{
		Path:        path,
		Total:       uint64(lpTotalNumberOfBytes),
		Free:        uint64(lpTotalNumberOfFreeBytes),
		Used:        uint64(lpTotalNumberOfBytes) - uint64(lpTotalNumberOfFreeBytes),
		UsedPercent: (float64(lpTotalNumberOfBytes) - float64(lpTotalNumberOfFreeBytes)) / float64(lpTotalNumberOfBytes) * 100,
		// InodesTotal: 0,
		// InodesFree: 0,
		// InodesUsed: 0,
		// InodesUsedPercent: 0,
	}
	return ret, nil
}

func Partitions(all bool) ([]PartitionStat, error) {
	return PartitionsWithContext(context.Background(), all)
}

func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	var ret []PartitionStat

	bufferSize := uint32(256)
	volnameBuffer := make([]byte, bufferSize)

	// find first volume
	handle, err := windows.FindFirstVolume((*uint16)(unsafe.Pointer(&volnameBuffer[0])), bufferSize)
	if nil != err {
		return ret, windows.GetLastError()
	}

	volumeInfo, err := getVolumeInfo(string(cleanVolumePath(volnameBuffer)))
	if err != nil {
		return ret, err
	}
	ret = append(ret, volumeInfo)

	// loop over all volumes, excluding the first volume
	for {
		// If no more partiotions, returns error, exit the loop
		err = windows.FindNextVolume(handle, (*uint16)(unsafe.Pointer(&volnameBuffer[0])), bufferSize)
		if err != nil {
			break
		}

		volumeInfo, err := getVolumeInfo(string(cleanVolumePath(volnameBuffer)))
		if err != nil {
			return ret, err
		}
		ret = append(ret, volumeInfo)
	}

	// Close the handle
	_ = windows.FindVolumeClose(handle)

	return ret, nil
}

func getVolumeInfo(volumeName string) (partition PartitionStat, err error) {
	d := PartitionStat{}

	bufferSize := uint32(256)
	volumeNameBuffer := make([]byte, bufferSize)
	lpFileSystemNameBuffer := make([]byte, bufferSize)
	volumeNameSerialNumber := uint32(0)
	maximumComponentLength := uint32(0)
	lpFileSystemFlags := uint32(0)

	volpath, _ := windows.UTF16PtrFromString(volumeName)
	typeret, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(volpath)))
	if typeret == 0 {
		return PartitionStat{}, windows.GetLastError()
	}
	// 2: DRIVE_REMOVABLE 3: DRIVE_FIXED 4: DRIVE_REMOTE 5: DRIVE_CDROM
	if typeret == 2 || typeret == 3 || typeret == 4 || typeret == 5 {
		err = windows.GetVolumeInformation(
			(*uint16)(unsafe.Pointer(volpath)),
			(*uint16)(unsafe.Pointer(&volumeNameBuffer[0])),
			bufferSize,
			&volumeNameSerialNumber,
			&maximumComponentLength,
			&lpFileSystemFlags,
			(*uint16)(unsafe.Pointer(&lpFileSystemNameBuffer[0])),
			bufferSize)

		if err != nil {
			if typeret == 5 || typeret == 2 {
				//device is not ready will happen if there is no disk in the drive
				return PartitionStat{}, errors.New("Device is not ready!")
			}
			return PartitionStat{}, windows.GetLastError()
		}

		opts := "rw"
		if lpFileSystemFlags&FileReadOnlyVolume != 0 {
			opts = "ro"
		}
		if lpFileSystemFlags&FileFileCompression != 0 {
			opts += ".compress"
		}

		d.Mountpoint = volumeName
		d.Device = string(bytes.Replace(volumeNameBuffer, []byte("\x00"), []byte(""), -1))
		d.Fstype = string(bytes.Replace(lpFileSystemNameBuffer, []byte("\x00"), []byte(""), -1))
		d.Opts = opts
	}

	return d, nil
}

func cleanVolumePath(data []byte) []byte {
	var res []byte
	for _, key := range data {
		if key != 0 {
			res = append(res, key)
		}
	}
	return res
}

func IOCounters(names ...string) (map[string]IOCountersStat, error) {
	return IOCountersWithContext(context.Background(), names...)
}

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	ret := make(map[string]IOCountersStat, 0)
	var dst []Win32_PerfFormattedData

	err := common.WMIQueryWithContext(ctx, "SELECT * FROM Win32_PerfFormattedData_PerfDisk_LogicalDisk", &dst)
	if err != nil {
		return ret, err
	}
	for _, d := range dst {
		if len(d.Name) > 3 { // not get _Total or Harddrive
			continue
		}

		if len(names) > 0 && !common.StringsHas(names, d.Name) {
			continue
		}

		ret[d.Name] = IOCountersStat{
			Name:       d.Name,
			ReadCount:  uint64(d.AvgDiskReadQueueLength),
			WriteCount: d.AvgDiskWriteQueueLength,
			ReadBytes:  uint64(d.AvgDiskBytesPerRead),
			WriteBytes: uint64(d.AvgDiskBytesPerWrite),
			ReadTime:   d.AvgDisksecPerRead,
			WriteTime:  d.AvgDisksecPerWrite,
		}
	}
	return ret, nil
}
