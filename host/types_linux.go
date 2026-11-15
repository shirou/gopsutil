// SPDX-License-Identifier: BSD-3-Clause
//go:build ignore

/*
Input to cgo -godefs.
*/

package host

/*
#include <sys/types.h>
#include <utmp.h>

typedef __typeof__(((struct utmp *)0)->ut_tv) ut_tv_t;

*/
import "C"

const sizeOfUtmp = C.sizeof_struct_utmp

type (
	utmp        C.struct_utmp
	exit_status C.struct_exit_status
	utTv        C.ut_tv_t
)
