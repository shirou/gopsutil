// SPDX-License-Identifier: BSD-3-Clause
//go:build ignore

// Hand writing: _Ctype_struct___0

/*
Input to cgo -godefs.
*/

package disk

/*
#include <sys/types.h>
#include <sys/statvfs.h>
#include <sys/cdefs.h>
#include <sys/featuretest.h>
#include <sys/stdint.h>
#include <machine/ansi.h>
#include <sys/ansi.h>

*/
import "C"

const (
	sizeOfStatvfs = C.sizeof_struct_statvfs
)

type (
	Statvfs C.struct_statvfs
)
