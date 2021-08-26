// +build linux

package so

import (
	"path/filepath"

	"golang.org/x/sys/unix"
)

const mntNSPath = "/ns/mnt"

type ns struct {
	dev uint64
	ino uint64
}

func getMntNS(pidPath string) ns {
	stat, _ := fstat(filepath.Join(pidPath, mntNSPath))
	return ns{stat.Dev, stat.Ino}
}

func fstat(path string) (stat unix.Stat_t, err error) {
	var fd int
	fd, err = unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return
	}
	defer unix.Close(fd)

	err = unix.Fstat(fd, &stat)
	return
}
