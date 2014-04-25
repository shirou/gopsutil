// +build linux
// +build amd64

package gopsutil

const (
	CLOCK_TICKS = 100  // C.sysconf(C._SC_CLK_TCK)
	PAGESIZE    = 4096 // C.sysconf(C._SC_PAGE_SIZE)
)
