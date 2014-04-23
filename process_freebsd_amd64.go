// +build freebsd
// +build amd64

package gopsutil

// copied from sys/sysctl.h
const (
	CTL_KERN           = 1  // "high kernel": proc, limits
	KERN_PROC          = 14 // struct: process entries
	KERN_PROC_PID      = 1  // by process id
	KERN_PROC_PROC     = 8  // only return procs
	KERN_PROC_PATHNAME = 12 // path to executable
)

// copied from sys/user.h
type Kinfo_proc struct {
	Ki_structsize   int32
	Ki_layout       int32
	Ki_args         int64
	Ki_paddr        int64
	Ki_addr         int64
	Ki_tracep       int64
	Ki_textvp       int64
	Ki_fd           int64
	Ki_vmspace      int64
	Ki_wchan        int64
	Ki_pid          int32
	Ki_ppid         int32
	Ki_pgid         int32
	Ki_tpgid        int32
	Ki_sid          int32
	Ki_tsid         int32
	Ki_jobc         [2]byte
	Ki_spare_short1 [2]byte
	Ki_tdev         int32
	Ki_siglist      [16]byte
	Ki_sigmask      [16]byte
	Ki_sigignore    [16]byte
	Ki_sigcatch     [16]byte
	Ki_uid          int32
	Ki_ruid         int32
	Ki_svuid        int32
	Ki_rgid         int32
	Ki_svgid        int32
	Ki_ngroups      [2]byte
	Ki_spare_short2 [2]byte
	Ki_groups       [64]byte
	Ki_size         int64
	Ki_rssize       int64
	Ki_swrss        int64
	Ki_tsize        int64
	Ki_dsize        int64
	Ki_ssize        int64
	Ki_xstat        [2]byte
	Ki_acflag       [2]byte
	Ki_pctcpu       int32
	Ki_estcpu       int32
	Ki_slptime      int32
	Ki_swtime       int32
	Ki_cow          int32
	Ki_runtime      int64
	Ki_start        [16]byte
	Ki_childtime    [16]byte
	Ki_flag         int64
	Ki_kiflag       int64
	Ki_traceflag    int32
	Ki_stat         [1]byte
	Ki_nice         [1]byte
	Ki_lock         [1]byte
	Ki_rqindex      [1]byte
	Ki_oncpu        [1]byte
	Ki_lastcpu      [1]byte
	Ki_ocomm        [17]byte
	Ki_wmesg        [9]byte
	Ki_login        [18]byte
	Ki_lockname     [9]byte
	Ki_comm         [20]byte
	Ki_emul         [17]byte
	Ki_sparestrings [68]byte
	Ki_spareints    [36]byte
	Ki_cr_flags     int32
	Ki_jid          int32
	Ki_numthreads   int32
	Ki_tid          int32
	Ki_pri          int32
	Ki_rusage       [144]byte
	Ki_rusage_ch    [144]byte
	Ki_pcb          int64
	Ki_kstack       int64
	Ki_udata        int64
	Ki_tdaddr       int64
	Ki_spareptrs    [48]byte
	Ki_spareint64s  [96]byte
	Ki_sflag        int64
	Ki_tdflags      int64
}
