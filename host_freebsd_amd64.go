// +build freebsd
// +build amd64

package gopsutil

const (
	UTNameSize = 16 /* see MAXLOGNAME in <sys/param.h> */
	UTLineSize = 8
	UTHostSize = 16
)

type utmp struct {
	UtLine [UTLineSize]byte
	UtName [UTNameSize]byte
	UtHost [UTHostSize]byte
	UtTime int32
}
