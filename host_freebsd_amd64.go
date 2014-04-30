// +build freebsd
// +build amd64

package gopsutil

const (
	UT_NAMESIZE = 16 /* see MAXLOGNAME in <sys/param.h> */
	UT_LINESIZE = 8
	UT_HOSTSIZE = 16
)

type utmp struct {
	UtLine [UT_LINESIZE]byte
	UtName [UT_NAMESIZE]byte
	UtHost [UT_HOSTSIZE]byte
	UtTime int32
}
