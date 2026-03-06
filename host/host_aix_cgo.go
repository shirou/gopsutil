// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && cgo

package host

/*
#include <procinfo.h>
*/
import "C"

import (
	"context"
	"unsafe"
)

func numProcs(_ context.Context) (uint64, error) {
	info := C.struct_procentry64{}
	cpid := C.pid_t(0)
	var count uint64
	for {
		n, err := C.getprocs64(unsafe.Pointer(&info), C.sizeof_struct_procentry64, nil, 0, &cpid, 1)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		count++
	}
	return count, nil
}
