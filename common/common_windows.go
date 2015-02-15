// +build windows

package common

import (
	"fmt"
	"os/exec"
	"strings"
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

// exec wmic and return lines splited by newline
func GetWmic(target string, query string) ([]string, error) {
	ret, err := exec.Command("wmic", target, "get", query, "/format:csv").Output()
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(ret), "\r\r\n")
	if len(lines) <= 2 {
		return []string{}, fmt.Errorf("wmic result malformed: [%q]", lines)
	}

	// skip first two line
	return lines[2:], nil
}
