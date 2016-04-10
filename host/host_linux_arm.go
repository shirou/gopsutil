// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_linux.go

package host

const (
	sizeofPtr	= 0x4
	sizeofShort	= 0x2
	sizeofInt	= 0x4
	sizeofLong	= 0x4
	sizeofLongLong	= 0x8
	sizeOfUtmp	= 0x180
)

type (
	_C_short	int16
	_C_int		int32
	_C_long		int32
	_C_long_long	int64
)

type utmp struct {
	Type			int16
	Pad_cgo_0		[2]byte
	Pid			int32
	Line			[32]uint8
	Id			[4]uint8
	User			[32]uint8
	Host			[256]uint8
	Exit			exit_status
	Session			int32
	Tv			timeval
	Addr_v6			[4]int32
	X__glibc_reserved	[20]uint8
}
type exit_status struct {
	Termination	int16
	Exit		int16
}
type timeval struct {
	Sec	int32
	Usec	int32
}
