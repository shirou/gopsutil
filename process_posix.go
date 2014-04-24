// +build linux freebsd

package gopsutil

import (
	"os"
	"strings"
	"syscall"
)

// POSIX
func getTerminalMap() (map[uint64]string, error) {
	ret := make(map[uint64]string)
	termfiles := make([]string, 0)

	d, err := os.Open("/dev")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	devnames, err := d.Readdirnames(-1)
	for _, devname := range devnames {
		if strings.HasPrefix(devname, "/dev/tty") {
			termfiles = append(termfiles, "/dev/tty/"+devname)
		}
	}

	ptsd, err := os.Open("/dev/pts")
	if err != nil {
		return nil, err
	}
	defer ptsd.Close()

	ptsnames, err := ptsd.Readdirnames(-1)
	for _, ptsname := range ptsnames {
		termfiles = append(termfiles, "/dev/pts/"+ptsname)
	}

	for _, name := range termfiles {
		stat := syscall.Stat_t{}
		syscall.Stat(name, &stat)
		rdev := uint64(stat.Rdev)
		ret[rdev] = strings.Replace(name, "/dev", "", -1)
	}
	return ret, nil
}
