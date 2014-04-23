// +build linux
// +build amd64

package gopsutil

type exit_status struct {
	E_termination int16 // Process termination status.
	E_exit        int16 // Process exit status.
}
type timeval struct {
	Tv_sec  uint32 // Seconds.
	Tv_usec uint32 // Microseconds.
}

type utmp struct {
	Ut_type    int16       // Type of login.
	Ut_pid     int32       // Process ID of login process.
	Ut_line    [32]byte    // Devicename.
	Ut_id      [4]byte     // Inittab ID.
	Ut_user    [32]byte    // Username.
	Ut_host    [256]byte   // Hostname for remote login.
	Ut_exit    exit_status // Exit status of a process marked
	Ut_session int32       // Session ID, used for windowing.
	Ut_tv      timeval     // Time entry was made.
	Ut_addr_v6 [16]byte    // Internet address of remote host.
	Unused     [20]byte    // Reserved for future use. // original is 20
}
