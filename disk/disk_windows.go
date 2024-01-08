//go:build windows
// +build windows

package disk

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/shirou/gopsutil/v3/internal/common"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const volumeNameBufferLength = uint32(windows.MAX_PATH + 1)
const volumePathBufferLength = volumeNameBufferLength

var (
	procGetDiskFreeSpaceExW              = common.Modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetLogicalDriveStringW           = common.Modkernel32.NewProc("GetLogicalDriveStringsW")
	procGetDriveTypeW                    = common.Modkernel32.NewProc("GetDriveTypeW")
	procGetVolumeInformationW            = common.Modkernel32.NewProc("GetVolumeInformationW")
	procFindFirstVolumeW                 = common.Modkernel32.NewProc("FindFirstVolumeW")
	procFindNextVolumeW                  = common.Modkernel32.NewProc("FindNextVolumeW")
	procFindVolumeClose                  = common.Modkernel32.NewProc("FindVolumeClose")
	procGetVolumePathNamesForVolumeNameW = common.Modkernel32.NewProc("GetVolumePathNamesForVolumeNameW")
)

var (
	fileFileCompression = int64(16)     // 0x00000010
	fileReadOnlyVolume  = int64(524288) // 0x00080000
)

// diskPerformance is an equivalent representation of DISK_PERFORMANCE in the Windows API.
// https://docs.microsoft.com/fr-fr/windows/win32/api/winioctl/ns-winioctl-disk_performance
type diskPerformance struct {
	BytesRead           int64
	BytesWritten        int64
	ReadTime            int64
	WriteTime           int64
	IdleTime            int64
	ReadCount           uint32
	WriteCount          uint32
	QueueDepth          uint32
	SplitCount          uint32
	QueryTime           int64
	StorageDeviceNumber uint32
	StorageManagerName  [8]uint16
	alignmentPadding    uint32 // necessary for 32bit support, see https://github.com/elastic/beats/pull/16553
}

func init() {
	// enable disk performance counters on Windows Server editions (needs to run as admin)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\PartMgr`, registry.SET_VALUE)
	if err == nil {
		key.SetDWordValue("EnableCounterForIoctl", 1)
		key.Close()
	}
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

// PartitionsWithContext returns disk partitions.
// Since GetVolumeInformation doesn't have a timeout, this method uses context to set deadline by users.
func PartitionsWithContext(ctx context.Context, all bool) ([]PartitionStat, error) {
	warnings := Warnings{
		Verbose: true,
	}

	var errFirstVol error
	retChan := make(chan PartitionStat)
	quitChan := make(chan struct{})
	defer close(quitChan)

	getPartitions := func() {
		defer close(retChan)

		volNameBuf := make([]uint16, volumeNameBufferLength)

		nextVolHandle, _, err := procFindFirstVolumeW.Call(
			uintptr(unsafe.Pointer(&volNameBuf[0])),
			uintptr(volumeNameBufferLength))
		if windows.Handle(nextVolHandle) == windows.InvalidHandle {
			errFirstVol = fmt.Errorf("failed to get first-volume: %w", err)
			return
		}
		defer procFindVolumeClose.Call(nextVolHandle)
		firstVolume := true
		var volPaths []string

		for {
			if !firstVolume {
				volNameBuf = make([]uint16, volumeNameBufferLength)
				if volRet, _, err := procFindNextVolumeW.Call(
					nextVolHandle,
					uintptr(unsafe.Pointer(&volNameBuf[0])),
					uintptr(volumeNameBufferLength)); err != nil && volRet == 0 {

					if errno, ok := err.(syscall.Errno); ok && errno == windows.ERROR_NO_MORE_FILES {
						break
					}
					warnings.Add(fmt.Errorf("failed to find next volume: %w", err))
					continue
				}
			}

			firstVolume = false
			if volPaths, err = getVolumePaths(volNameBuf); err != nil {
				warnings.Add(fmt.Errorf("failed to find paths for volume %s", windows.UTF16ToString(volNameBuf)))
				continue
			}

			if len(volPaths) > 0 {
				volPathPtr, _ := windows.UTF16PtrFromString(volPaths[0])
				driveType, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(volPathPtr)))
				if driveType == windows.DRIVE_UNKNOWN {
					err := windows.GetLastError()
					warnings.Add(err)
					continue
				}
				for _, volPath := range volPaths {
					if driveType == windows.DRIVE_REMOVABLE || driveType == windows.DRIVE_FIXED || driveType == windows.DRIVE_REMOTE || driveType == windows.DRIVE_CDROM {
						fsFlags, fsNameBuf := uint32(0), make([]uint16, 256)
						rootPathPtr, _ := windows.UTF16PtrFromString(volPath)
						volNameBuf := make([]uint16, 256)
						volSerialNum := uint32(0)
						maxComponentLen := uint32(0)
						driveRet, _, err := procGetVolumeInformationW.Call(
							uintptr(unsafe.Pointer(rootPathPtr)),
							uintptr(unsafe.Pointer(&volNameBuf[0])),
							uintptr(len(volNameBuf)),
							uintptr(unsafe.Pointer(&volSerialNum)),
							uintptr(unsafe.Pointer(&maxComponentLen)),
							uintptr(unsafe.Pointer(&fsFlags)),
							uintptr(unsafe.Pointer(&fsNameBuf[0])),
							uintptr(len(fsNameBuf)))
						if err != nil && driveRet == 0 {
							if driveType == windows.DRIVE_CDROM || driveType == windows.DRIVE_REMOVABLE {
								continue
							}
							warnings.Add(fmt.Errorf("failed to get volume information: %w", err))
							continue
						}
						opts := []string{"rw"}
						if int64(fsFlags)&fileReadOnlyVolume != 0 {
							opts = []string{"ro"}
						}
						if int64(fsFlags)&fileFileCompression != 0 {
							opts = append(opts, "compress")
						}

						path := strings.TrimRight(volPath, "\\")

						select {
						case retChan <- PartitionStat{
							Mountpoint: path,
							Device:     path,
							Fstype:     windows.UTF16PtrToString(&fsNameBuf[0]),
							Opts:       opts,
						}:
						case <-quitChan:
							return
						}
					}
				}
			}
		}
	}

	go getPartitions()

	var ret []PartitionStat
	for {
		select {
		case p, ok := <-retChan:
			if !ok {
				if errFirstVol != nil {
					return ret, errFirstVol
				}
				return ret, warnings.Reference()
			}
			ret = append(ret, p)
		case <-ctx.Done():
			return ret, ctx.Err()
		}
	}
}

// getVolumePaths returns the path for the given volume name.
func getVolumePaths(volNameBuf []uint16) ([]string, error) {
	volPathsBuf := make([]uint16, volumePathBufferLength)
	returnLen := uint32(0)
	if result, _, err := procGetVolumePathNamesForVolumeNameW.Call(
		uintptr(unsafe.Pointer(&volNameBuf[0])),
		uintptr(unsafe.Pointer(&volPathsBuf[0])),
		uintptr(volumePathBufferLength),
		uintptr(unsafe.Pointer(&returnLen))); err != nil && result == 0 {
		return nil, err
	}
	return split0(volPathsBuf, int(returnLen)), nil
}

// split0 iterates through s16 upto `end` and slices `s16` into sub-slices separated by the null character (uint16(0)).
// split0 converts the sub-slices between the null characters into strings then returns them in a slice.
func split0(s16 []uint16, end int) []string {
	if end > len(s16) {
		end = len(s16)
	}

	from, ss := 0, make([]string, 0)

	for to := 0; to < end; to++ {
		if s16[to] == 0 {
			if from < to && s16[from] != 0 {
				ss = append(ss, string(utf16.Decode(s16[from:to])))
			}
			from = to + 1
		}
	}

	return ss
}

func IOCountersWithContext(ctx context.Context, names ...string) (map[string]IOCountersStat, error) {
	// https://github.com/giampaolo/psutil/blob/544e9daa4f66a9f80d7bf6c7886d693ee42f0a13/psutil/arch/windows/disk.c#L83
	drivemap := make(map[string]IOCountersStat, 0)
	var diskPerformance diskPerformance

	lpBuffer := make([]uint16, 254)
	lpBufferLen, err := windows.GetLogicalDriveStrings(uint32(len(lpBuffer)), &lpBuffer[0])
	if err != nil {
		return drivemap, err
	}
	for _, v := range lpBuffer[:lpBufferLen] {
		if 'A' <= v && v <= 'Z' {
			path := string(rune(v)) + ":"
			typepath, _ := windows.UTF16PtrFromString(path)
			typeret := windows.GetDriveType(typepath)
			if typeret == 0 {
				return drivemap, windows.GetLastError()
			}
			if typeret != windows.DRIVE_FIXED {
				continue
			}
			szDevice := fmt.Sprintf(`\\.\%s`, path)
			const IOCTL_DISK_PERFORMANCE = 0x70020
			h, err := windows.CreateFile(syscall.StringToUTF16Ptr(szDevice), 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
			if err != nil {
				if err == windows.ERROR_FILE_NOT_FOUND {
					continue
				}
				return drivemap, err
			}
			defer windows.CloseHandle(h)

			var diskPerformanceSize uint32
			err = windows.DeviceIoControl(h, IOCTL_DISK_PERFORMANCE, nil, 0, (*byte)(unsafe.Pointer(&diskPerformance)), uint32(unsafe.Sizeof(diskPerformance)), &diskPerformanceSize, nil)
			if err != nil {
				return drivemap, err
			}
			drivemap[path] = IOCountersStat{
				ReadBytes:  uint64(diskPerformance.BytesRead),
				WriteBytes: uint64(diskPerformance.BytesWritten),
				ReadCount:  uint64(diskPerformance.ReadCount),
				WriteCount: uint64(diskPerformance.WriteCount),
				ReadTime:   uint64(diskPerformance.ReadTime / 10000 / 1000), // convert to ms: https://github.com/giampaolo/psutil/issues/1012
				WriteTime:  uint64(diskPerformance.WriteTime / 10000 / 1000),
				Name:       path,
			}
		}
	}
	return drivemap, nil
}

func SerialNumberWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(ctx context.Context, name string) (string, error) {
	return "", common.ErrNotImplementedError
}
