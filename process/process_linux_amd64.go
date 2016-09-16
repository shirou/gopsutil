// +build linux
// +build amd64

package process

const (
	ClockTicks = 100 // C.sysconf(C._SC_CLK_TCK)
)
