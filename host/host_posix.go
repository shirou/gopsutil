// +build linux freebsd openbsd darwin solaris

package host

import (
	"golang.org/x/sys/unix"
)

func kernelArch() (string, error) {
	var utsname unix.Utsname
	err := unix.Uname(&utsname)
	return string(utsname.Machine[:]), err
}
