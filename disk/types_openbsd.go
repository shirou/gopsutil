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
*/
import "C"

const (
	DEVSTAT_NO_DATA = 0x00
	DEVSTAT_READ    = 0x01
	DEVSTAT_WRITE   = 0x02
	DEVSTAT_FREE    = 0x03
)

const (
	sizeOfDiskstats = C.sizeof_struct_diskstats
)

type Diskstats C.struct_diskstats
type Timeval C.struct_timeval

type Diskstat C.struct_diskstat
type Bintime C.struct_bintime
