// +build ignore
// Hand writing: _Ctype_struct___0

/*
Input to cgo -godefs.

*/

package disk

/*
#include <sys/types.h>
#include <sys/mount.h>
#include <devstat.h>

enum {
	sizeofPtr = sizeof(void*),
};

// because statinfo has long double snap_time, redefine with changing long long
struct statinfo2 {
        long            cp_time[CPUSTATES];
        long            tk_nin;
        long            tk_nout;
        struct devinfo  *dinfo;
        long long       snap_time;
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
	sizeOfDevstat = C.sizeof_struct_devstat
)

// Basic types

type (
	_C_short       C.short
	_C_int         C.int
	_C_long        C.long
	_C_long_long   C.longlong
	_C_long_double C.longlong
)

type Devstat C.struct_devstat
type Bintime C.struct_bintime
