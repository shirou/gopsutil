// +build linux

// This file is copied from elastic/gosigar
// https://github.com/elastic/gosigar/tree/master/sys/linux

package netlink

import (
	"encoding/binary"
	"unsafe"
)

func GetEndian() binary.ByteOrder {
	var i int32 = 0x1
	v := (*[4]byte)(unsafe.Pointer(&i))
	if v[0] == 0 {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}
