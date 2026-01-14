// SPDX-License-Identifier: BSD-3-Clause
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs types_darwin.go

package host

type utmpx32 struct {
	User [256]int8
	Id   [4]int8
	Line [32]int8
	Pid  int32
	Type int16
	Tv   timeval32
	Host [256]int8
	Pad  [16]uint32
}

type timeval32 struct {
	Sec  int32
	Usec int32
}
