// +build windows

package main

import (
	"syscall"
)

var (
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
)
