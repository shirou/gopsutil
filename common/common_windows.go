// +build windows

package common

import (
	"syscall"
	"unsafe"
)

var (
	Modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	ModNt       = syscall.NewLazyDLL("ntdll.dll")

	ProcGetSystemTimes           = Modkernel32.NewProc("GetSystemTimes")
	ProcNtQuerySystemInformation = ModNt.NewProc("NtQuerySystemInformation")
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

// borrowed from net/interface_windows.go
func BytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}
