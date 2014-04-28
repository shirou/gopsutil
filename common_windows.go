// +build windows

package gopsutil

import (
	"syscall"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modNt       = syscall.NewLazyDLL("ntdll.dll")

	procGetSystemTimes           = modkernel32.NewProc("GetSystemTimes")
	procNtQuerySystemInformation = modNt.NewProc("NtQuerySystemInformation")
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}
