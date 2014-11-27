// +build linux
// +build arm

package gopsutil

type exitStatus struct {
	Etermination int16 // Process termination status.
	Eexit        int16 // Process exit status.
}
type timeval struct {
	TvSec  uint32 // Seconds.
	TvUsec uint32 // Microseconds.
}

type utmp struct {
	UtType    int16      // Type of login.
	UtPid     int32      // Process ID of login process.
	UtLine    [32]byte   // Devicename.
	UtID      [4]byte    // Inittab ID.
	UtUser    [32]byte   // Username.
	UtHost    [256]byte  // Hostname for remote login.
	UtExit    exitStatus // Exit status of a process marked
	UtSession int32      // Session ID, used for windowing.
	UtTv      timeval    // Time entry was made.
	UtAddrV6  [16]byte   // Internet address of remote host.
	Unused    [20]byte   // Reserved for future use. // original is 20
}
