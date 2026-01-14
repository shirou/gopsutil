// SPDX-License-Identifier: BSD-3-Clause
//go:build ignore

// plus hand editing about timeval

/*
Input to cgo -godefs.
*/

package host

/*
#include <utmpx.h>
#include <sys/_types/_timeval32.h>

// https://github.com/apple-oss-distributions/Libc/blob/55b54c0a0c37b3b24393b42b90a4c561d6c606b1/gen/utmpx-darwin.h#L86
struct utmpx32 {
  char ut_user[_UTX_USERSIZE];
  char ut_id[_UTX_IDSIZE];
  char ut_line[_UTX_LINESIZE];
  pid_t ut_pid;
  short ut_type;
  struct timeval32 ut_tv;
  char ut_host[_UTX_HOSTSIZE];
  __uint32_t ut_pad[16];
};
*/
import "C"

type (
	utmpx32   C.struct_utmpx32
	timeval32 C.struct_timeval32
)
