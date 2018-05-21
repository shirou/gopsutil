// +build windows

package winservices

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modadvapi32              = windows.NewLazySystemDLL("advapi32.dll")
	procQueryServiceStatusEx = modadvapi32.NewProc("QueryServiceStatusEx")
)

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

// SC_STATUS_PROCESS_INFO reference to https://msdn.microsoft.com/en-us/library/windows/desktop/ms684941(v=vs.85).aspx,
// Use SC_STATUS_PROCESS_INFO to retrieve the service status information.
const SC_STATUS_PROCESS_INFO int = 0

// QueryServiceStatusEx golang/sys/windows standard library can not implement QueryServiceStatusEx.
func QueryServiceStatusEx(service windows.Handle, infoLevel int, buff *byte, buffSize uint32, bytesNeeded *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procQueryServiceStatusEx.Addr(), 5, uintptr(service), uintptr(infoLevel), uintptr(unsafe.Pointer(buff)), uintptr(buffSize), uintptr(unsafe.Pointer(bytesNeeded)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
