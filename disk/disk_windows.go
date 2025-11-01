// SPDX-License-Identifier: BSD-3-Clause
//go:build windows

package disk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/shirou/gopsutil/v4/internal/common"
)

const (
	maxVolumeNameLength     = uint32(windows.MAX_PATH + 1) // this should be a GUID (50), but for safety I keep max_path
	maxFileSystemNameLength = uint32(windows.MAX_PATH + 1)
	maxWarningsInDrive      = 5
)

// https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getvolumeinformationa#parameters
const (
	rw       = "rw"
	ro       = "ro"
	compress = "compress"
)

const (
	firstPossibleDriveLetter = 'A'
	lastPossibleDriveLetter  = 'Z'
)

var (
	procGetDiskFreeSpaceExW              = common.Modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetLogicalDriveStringsW          = common.Modkernel32.NewProc("GetLogicalDriveStringsW")
	procGetVolumeInformation             = common.Modkernel32.NewProc("GetVolumeInformationW")
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

func UsageWithContext(_ context.Context, path string) (*UsageStat, error) {
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
// It uses procGetLogicalDriveStringsW to get drives with drive letters and procFindFirstVolumeW to get volumes without drive letters.
// Since the api calls don't have a timeout, this method uses context to set deadline by users.
func PartitionsWithContext(_ context.Context, _ bool) ([]PartitionStat, error) {
	warnings := Warnings{Verbose: true}
	processedPaths := make(map[string]struct{})
	partitionStats := []PartitionStat{}

	// Get drives with drive letters (including remote drives, ex: SMB shares)
	drives, err := getLogicalDrives()
	if err != nil {
		return partitionStats, err
	}

	partitionStats = processLogicalDrives(drives, processedPaths, partitionStats, warnings)

	// Get volumes without drive letters (ex: mounted folders with no drive letter)
	partitionStats = processVolumesMountedAsFolders(partitionStats, warnings, processedPaths)
	return partitionStats, warnings.Reference()
}

func processVolumesMountedAsFolders(partitionStats []PartitionStat, warnings Warnings, processedPaths map[string]struct{}) []PartitionStat {
	volNameBuf := make([]uint16, maxVolumeNameLength)
	nextVolHandle, _, err := procFindFirstVolumeW.Call(
		uintptr(unsafe.Pointer(&volNameBuf[0])),
		uintptr(maxVolumeNameLength))
	if windows.Handle(nextVolHandle) == windows.InvalidHandle {
		warnings.Add(fmt.Errorf("failed to get first-volume: %w", err))
		return partitionStats
	}
	defer procFindVolumeClose.Call(nextVolHandle)
	for {
		mounts, err := getVolumePaths(volNameBuf)
		if err != nil {
			warnings.Add(fmt.Errorf("failed to find paths for volume %s", windows.UTF16ToString(volNameBuf)))
			continue
		}

		for _, mount := range mounts {
			if _, ok := processedPaths[mount]; ok {
				continue
			}
			if partitionStat, warning := buildPartitionStat(mount); warning == nil {
				partitionStats = append(partitionStats, partitionStat)
			} else {
				warnings.Add(warning)
			}
		}

		volNameBuf = make([]uint16, maxVolumeNameLength)
		if volRet, _, err := procFindNextVolumeW.Call(
			nextVolHandle,
			uintptr(unsafe.Pointer(&volNameBuf[0])),
			uintptr(maxVolumeNameLength)); err != nil && volRet == 0 {
			var errno syscall.Errno
			if errors.As(err, &errno) && errno == windows.ERROR_NO_MORE_FILES {
				break
			}
			warnings.Add(fmt.Errorf("failed to find next volume: %w", err))
			if len(warnings.List) > maxWarningsInDrive {
				break
			}
		}
	}
	return partitionStats
}

func processLogicalDrives(drives []string, processedPaths map[string]struct{}, partitionStats []PartitionStat, warnings Warnings) []PartitionStat {
	for _, drive := range drives {
		if drive != "" && drive[0] >= firstPossibleDriveLetter && drive[0] <= lastPossibleDriveLetter {
			v := drive[0]
			path := string(v) + ":"
			if partitionStat, warning := buildPartitionStat(path); warning == nil {
				processedPaths[partitionStat.Mountpoint+"\\"] = struct{}{}
				partitionStats = append(partitionStats, partitionStat)
			} else {
				warnings.Add(warning)
			}
		}
	}
	return partitionStats
}

// getLogicalDrives retrieves all logical drives using GetLogicalDriveStringsW.
// We first call GetLogicalDriveStringsW with a buffer length of 0 to get the required buffer size.
// https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getlogicaldrivestringsw
func getLogicalDrives() ([]string, error) {
	bufferLen, _, err := procGetLogicalDriveStringsW.Call(
		uintptr(0),
		uintptr(0))
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err // The call failed with an unexpected error
	}
	lpBuffer := make([]uint16, bufferLen)
	// buffer can be longer than MAX_PATH
	_, _, err = procGetLogicalDriveStringsW.Call(
		uintptr(len(lpBuffer)),
		uintptr(unsafe.Pointer(&lpBuffer[0])))
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil, err // The call failed with an unexpected error
	}

	drivesString := windows.UTF16ToString(lpBuffer)
	drives := strings.Split(drivesString, "\x00")
	return drives, nil
}

func buildPartitionStat(path string) (PartitionStat, error) {
	typePath, _ := windows.UTF16PtrFromString(path)
	driveType := windows.GetDriveType(typePath)

	if driveType == windows.DRIVE_UNKNOWN {
		return PartitionStat{}, windows.GetLastError()
	}

	if driveType == windows.DRIVE_REMOVABLE || driveType == windows.DRIVE_FIXED ||
		driveType == windows.DRIVE_REMOTE || driveType == windows.DRIVE_CDROM {
		volPath, _ := windows.UTF16PtrFromString(path + "/")
		volumeName := make([]byte, maxVolumeNameLength)
		fsName := make([]byte, maxFileSystemNameLength)
		var fsFlags int64

		// https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getvolumeinformationw
		ret, _, err := procGetVolumeInformation.Call(
			uintptr(unsafe.Pointer(volPath)),
			uintptr(unsafe.Pointer(&volumeName[0])),
			uintptr(len(volumeName)),
			uintptr(0), // serial number
			uintptr(0), // max component length
			uintptr(unsafe.Pointer(&fsFlags)),
			uintptr(unsafe.Pointer(&fsName[0])),
			uintptr(len(fsName)),
		)

		if ret == 0 {
			if driveType == windows.DRIVE_REMOVABLE || driveType == windows.DRIVE_REMOTE || driveType == windows.DRIVE_CDROM {
				return PartitionStat{}, nil // Device not ready
			}
			return PartitionStat{}, err
		}

		opts := []string{rw}
		if fsFlags&fileReadOnlyVolume != 0 {
			opts = []string{ro}
		}
		if fsFlags&fileFileCompression != 0 {
			opts = append(opts, compress)
		}

		return PartitionStat{
			Mountpoint: path,
			Device:     path,
			Fstype:     string(bytes.ReplaceAll(fsName, []byte("\x00"), []byte(""))),
			Opts:       opts,
		}, nil
	}

	return PartitionStat{}, nil
}

func IOCountersWithContext(_ context.Context, names ...string) (map[string]IOCountersStat, error) {
	// https://github.com/giampaolo/psutil/blob/544e9daa4f66a9f80d7bf6c7886d693ee42f0a13/psutil/arch/windows/disk.c#L83
	drivemap := make(map[string]IOCountersStat, 0)
	var dPerformance diskPerformance

	lpBuffer := make([]uint16, 254)
	lpBufferLen, err := windows.GetLogicalDriveStrings(uint32(len(lpBuffer)), &lpBuffer[0])
	if err != nil {
		return drivemap, err
	}
	for _, v := range lpBuffer[:lpBufferLen] {
		if v < firstPossibleDriveLetter || v > lastPossibleDriveLetter {
			continue
		}
		path := string(rune(v)) + ":"
		typepath, _ := windows.UTF16PtrFromString(path)
		typeret := windows.GetDriveType(typepath)
		if typeret != windows.DRIVE_FIXED {
			continue
		}
		szDevice := `\\.\` + path
		const IOCTL_DISK_PERFORMANCE = 0x70020
		h, err := windows.CreateFile(windows.StringToUTF16Ptr(szDevice), 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
		if err != nil {
			if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
				continue
			}
			return drivemap, err
		}
		defer windows.CloseHandle(h)

		var diskPerformanceSize uint32
		err = windows.DeviceIoControl(h, IOCTL_DISK_PERFORMANCE, nil, 0, (*byte)(unsafe.Pointer(&dPerformance)), uint32(unsafe.Sizeof(dPerformance)), &diskPerformanceSize, nil)
		if err != nil {
			return drivemap, err
		}

		if len(names) == 0 || common.StringsHas(names, path) {
			drivemap[path] = IOCountersStat{
				ReadBytes:  uint64(dPerformance.BytesRead),
				WriteBytes: uint64(dPerformance.BytesWritten),
				ReadCount:  uint64(dPerformance.ReadCount),
				WriteCount: uint64(dPerformance.WriteCount),
				ReadTime:   uint64(dPerformance.ReadTime / 10000 / 1000), // convert to ms: https://github.com/giampaolo/psutil/issues/1012
				WriteTime:  uint64(dPerformance.WriteTime / 10000 / 1000),
				Name:       path,
			}
		}
	}
	return drivemap, nil
}

func SerialNumberWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}

func LabelWithContext(_ context.Context, _ string) (string, error) {
	return "", common.ErrNotImplementedError
}

// getVolumePaths returns the path for the given volume name.
func getVolumePaths(volNameBuf []uint16) ([]string, error) {
	// https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getvolumepathnamesforvolumenamew
	// Consider that NTFS supports paths longer than windows.MAX_PATH
	returnLen := uint32(0)
	firstResult, _, volumePathFirstErr := procGetVolumePathNamesForVolumeNameW.Call(
		uintptr(unsafe.Pointer(&volNameBuf[0])),
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(&returnLen)))
	if firstResult == 0 && !errors.Is(volumePathFirstErr, windows.ERROR_MORE_DATA) {
		return nil, fmt.Errorf("failed to get volume paths size for volume %s: %w", windows.UTF16ToString(volNameBuf), volumePathFirstErr)
	}

	volPathsBuf := make([]uint16, returnLen)
	ok, _, volumePathNamesErr := procGetVolumePathNamesForVolumeNameW.Call(
		uintptr(unsafe.Pointer(&volNameBuf[0])),
		uintptr(unsafe.Pointer(&volPathsBuf[0])),
		uintptr(returnLen),
		uintptr(unsafe.Pointer(&returnLen)))
	if ok != 0 {
		return split0(volPathsBuf, int(returnLen)), nil
	}
	return nil, fmt.Errorf("failed to get volume paths for volume %s: %w", windows.UTF16ToString(volNameBuf), volumePathNamesErr)
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
