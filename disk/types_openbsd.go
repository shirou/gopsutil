// +build ignore
// Hand writing: _Ctype_struct___0

/*
Input to cgo -godefs.
*/

package disk

/*
#include <sys/types.h>
#include <sys/disk.h>
#include <sys/mount.h>

enum {
	sizeofPtr = sizeof(void*),
};

*/
import "C"

// Machine characteristics; for internal use.

const (
	sizeofPtr        = C.sizeofPtr
	sizeofShort      = C.sizeof_short
	sizeofInt        = C.sizeof_int
	sizeofLong       = C.sizeof_long
	sizeofLongLong   = C.sizeof_longlong
	sizeofLongDouble = C.sizeof_longlong

	DEVSTAT_NO_DATA = 0x00
	DEVSTAT_READ    = 0x01
	DEVSTAT_WRITE   = 0x02
	DEVSTAT_FREE    = 0x03
)

const (
	sizeOfDiskstats = C.sizeof_struct_diskstats
)

// Basic types

type (
	_C_short       C.short
	_C_int         C.int
	_C_long        C.long
	_C_long_long   C.longlong
	_C_long_double C.longlong
)

type Statfs C.struct_statfs
type Diskstats C.struct_diskstats
type Fsid C.fsid_t
type Timeval C.struct_timeval

type Diskstat C.struct_diskstat
type Bintime C.struct_bintime
