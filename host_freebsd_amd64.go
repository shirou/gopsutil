// +build freebsd
// +build amd64

package gopsutil

const (
	UT_NAMESIZE = 16 /* see MAXLOGNAME in <sys/param.h> */
	UT_LINESIZE = 8
	UT_HOSTSIZE = 16
)

type utmp struct {
	Ut_line [UT_LINESIZE]byte
	Ut_name [UT_NAMESIZE]byte
	Ut_host [UT_HOSTSIZE]byte
	Ut_time int32
}
