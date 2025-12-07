// SPDX-License-Identifier: BSD-3-Clause
//go:build openbsd

package common

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func CallSyscall(mib []int32) (buf []byte, length uint64, err error) {
	mibptr := unsafe.Pointer(&mib[0])
	miblen := uint64(len(mib))

	// get required buffer size
	length = uint64(0)
	_, _, errno := unix.Syscall6(
		unix.SYS___SYSCTL,
		uintptr(mibptr),
		uintptr(miblen),
		0,
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if errno != 0 {
		return nil, length, errno
	}
	if length == 0 {
		return nil, length, nil
	}
	// get proc info itself
	buf = make([]byte, length)
	_, _, errno = unix.Syscall6(
		unix.SYS___SYSCTL,
		uintptr(mibptr),
		uintptr(miblen),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if errno != 0 {
		return buf, length, errno
	}
	return buf, length, nil
}
