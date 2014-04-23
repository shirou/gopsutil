// +build linux,amd64

package gopsutil

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"syscall"
	"unsafe"
)

func HostInfo() (HostInfoStat, error) {
	ret := HostInfoStat{}

	hostname, err := os.Hostname()
	ret.Hostname = hostname
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func Boot_time() (int64, error) {
	sysinfo := &syscall.Sysinfo_t{}
	if err := syscall.Sysinfo(sysinfo); err != nil {
		return 0, err
	}
	return sysinfo.Uptime, nil
}

func Users() ([]UserStat, error) {
	utmpfile := "/var/run/utmp"
	ret := make([]UserStat, 0)

	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	u := utmp{}
	entrySize := int(unsafe.Sizeof(u))
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]

		var u utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil {
			continue
		}
		user := UserStat{
			User:     byteToString(u.Ut_user[:]),
			Terminal: byteToString(u.Ut_line[:]),
			Host:     byteToString(u.Ut_host[:]),
			Started:  int(u.Ut_tv.Tv_sec),
		}
		ret = append(ret, user)
	}

	return ret, nil

}
