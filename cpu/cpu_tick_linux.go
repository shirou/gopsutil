// +build linux

package cpu

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

// jiffies ( getconf CLK_TCK )
func GetClockTick() uint {
	var sc_clk_tck C.long
	sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
	return uint(sc_clk_tck)
}
