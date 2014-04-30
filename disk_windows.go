// +build windows

package gopsutil

import (
	"bytes"
	"errors"
	"syscall"
	"unsafe"
)

var (
	procGetDiskFreeSpaceExW     = modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetLogicalDriveStringsW = modkernel32.NewProc("GetLogicalDriveStringsW")
	procGetDriveType            = modkernel32.NewProc("GetDriveTypeW")
	provGetVolumeInformation    = modkernel32.NewProc("GetVolumeInformationW")
)

var (
	FILE_FILE_COMPRESSION = int64(16)     // 0x00000010
	FILE_READ_ONLY_VOLUME = int64(524288) // 0x00080000
)

func DiskUsage(path string) (DiskUsageStat, error) {
	ret := DiskUsageStat{}

	ret.Path = path
	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	diskret, _, err := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)))
	if diskret == 0 {
		return ret, err
	}
	ret.Total = uint64(lpTotalNumberOfBytes)
	//	ret.Free = uint64(lpFreeBytesAvailable) // python psutil does not use this
	ret.Free = uint64(lpTotalNumberOfFreeBytes)
	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Used) / float64(ret.Total) * 100.0

	return ret, nil
}

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat
	lpBuffer := make([]byte, 254)
	diskret, _, err := procGetLogicalDriveStringsW.Call(
		uintptr(len(lpBuffer)),
		uintptr(unsafe.Pointer(&lpBuffer[0])))
	if diskret == 0 {
		return ret, err
	}
	for _, v := range lpBuffer {
		if v >= 65 && v <= 90 {
			path := string(v) + ":"
			if path == "A:" || path == "B:" { // skip floppy drives
				continue
			}
			typepath, _ := syscall.UTF16PtrFromString(path)
			typeret, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(typepath)))
			if typeret == 0 {
				return ret, syscall.GetLastError()
			}
			// 2: DRIVE_REMOVABLE 3: DRIVE_FIXED 5: DRIVE_CDROM

			if typeret == 2 || typeret == 3 || typeret == 5 {
				lpVolumeNameBuffer := make([]byte, 256)
				lpVolumeSerialNumber := int64(0)
				lpMaximumComponentLength := int64(0)
				lpFileSystemFlags := int64(0)
				lpFileSystemNameBuffer := make([]byte, 256)
				volpath, _ := syscall.UTF16PtrFromString(string(v) + ":/")
				driveret, _, err := provGetVolumeInformation.Call(
					uintptr(unsafe.Pointer(volpath)),
					uintptr(unsafe.Pointer(&lpVolumeNameBuffer[0])),
					uintptr(len(lpVolumeNameBuffer)),
					uintptr(unsafe.Pointer(&lpVolumeSerialNumber)),
					uintptr(unsafe.Pointer(&lpMaximumComponentLength)),
					uintptr(unsafe.Pointer(&lpFileSystemFlags)),
					uintptr(unsafe.Pointer(&lpFileSystemNameBuffer[0])),
					uintptr(len(lpFileSystemNameBuffer)))
				if driveret == 0 {
					return ret, err
				}
				opts := "rw"
				if lpFileSystemFlags&FILE_READ_ONLY_VOLUME != 0 {
					opts = "ro"
				}
				if lpFileSystemFlags&FILE_FILE_COMPRESSION != 0 {
					opts += ".compress"
				}

				d := DiskPartitionStat{
					Mountpoint: path,
					Device:     path,
					Fstype:     string(bytes.Replace(lpFileSystemNameBuffer, []byte("\x00"), []byte(""), -1)),
					Opts:       opts,
				}
				ret = append(ret, d)
			}
		}
	}
	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	ret := make(map[string]DiskIOCountersStat, 0)
	return ret, errors.New("not implemented yet")
}
