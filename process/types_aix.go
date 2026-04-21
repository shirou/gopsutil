// SPDX-License-Identifier: BSD-3-Clause
//go:build ignore

// Input to cgo -godefs. See mktypes.sh in the root directory.
// go tool cgo -godefs types_aix.go | sed 's/\*byte/uint64/' > process_aix_ppc64.go

package process

/*
#include <sys/procfs.h>
*/
import "C"

type prTimestruc64 C.struct_pr_timestruc64
type lwpSinfo C.struct_lwpsinfo
type psinfo C.struct_psinfo
