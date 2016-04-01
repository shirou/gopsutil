// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_freebsd.go

package host

const (
	sizeofPtr      = 0x8
	sizeofShort    = 0x2
	sizeofInt      = 0x4
	sizeofLong     = 0x8
	sizeofLongLong = 0x8
)

type (
	_C_short     int16
	_C_int       int32
	_C_long      int64
	_C_long_long int64
)

type Utmp struct {
	Line [8]int8
	Name [16]int8
	Host [16]int8
	Time int32
}
type Utmpx struct {
	Type        int16
	Pad_cgo_0   [6]byte
	Tv          Timeval
	Id          [8]int8
	Pid         int32
	User        [32]int8
	Line        [16]int8
	Host        [128]int8
	X__ut_spare [64]int8
	Pad_cgo_1   [4]byte
}
type Timeval struct {
	Sec  int64
	Usec int64
}
