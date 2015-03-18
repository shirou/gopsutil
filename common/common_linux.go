// +build linux
package common

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

/**
C system calls to go types
*/

/**
Get clock ticks per second from linux
*/
func GetClockTicksPerSecond() (ticksPerSecond uint64) {
	var sc_clk_tck C.long
	sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
	ticksPerSecond = uint64(sc_clk_tck)
	return
}
