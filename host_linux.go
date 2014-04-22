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

type exit_status struct {
	E_termination int16 // Process termination status.
	E_exit        int16 // Process exit status.
}
type timeval struct {
	Tv_sec  uint32 // Seconds.
	Tv_usec uint32 // Microseconds.
}

type utmp struct {
	Ut_type    int16       // Type of login.
	Ut_pid     int32       // Process ID of login process.
	Ut_line    [32]byte    // Devicename.
	Ut_id      [4]byte     // Inittab ID.
	Ut_user    [32]byte    // Username.
	Ut_host    [256]byte   // Hostname for remote login.
	Ut_exit    exit_status // Exit status of a process marked
	Ut_session int32       // Session ID, used for windowing.
	Ut_tv      timeval     // Time entry was made.
	Ut_addr_v6 [16]byte    // Internet address of remote host.
	Unused     [20]byte    // Reserved for future use. // original is 20
}

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
