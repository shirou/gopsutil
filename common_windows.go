// +build windows

package gopsutil

import (
	"syscall"
)

var (
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
)
