// +build freebsd

package gopsutil

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
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

func BootTime() (int64, error) {
	values, err := doSysctrl("kern.boottime")
	if err != nil {
		return 0, err
	}
	// ex: { sec = 1392261637, usec = 627534 } Thu Feb 13 12:20:37 2014
	v := strings.Replace(values[2], ",", "", 1)

	boottime, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return boottime, nil
}

func Users() ([]UserStat, error) {
	utmpfile := "/var/run/utmp"
	var ret []UserStat

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
			User:     byteToString(u.Ut_name[:]),
			Terminal: byteToString(u.Ut_line[:]),
			Host:     byteToString(u.Ut_host[:]),
			Started:  int(u.Ut_time),
		}
		ret = append(ret, user)
	}

	return ret, nil

}
