// +build windows

package gopsutil

import (
	"syscall"
)

var (
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
)

type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}
