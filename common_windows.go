// +build windows

package gopsutil

import (
	"syscall"
	"unsafe"
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

// borrowed from net/interface_windows.go
func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}
