package cpu

import "time"

//go:generate easyjson -build_tags solaris -output_filename cpu_json_illumos.generated.go cpu_json_illumos.go

// TimeZoneStat is the output from the "zones" kstat(1M) module
//easyjson:json
type TimesZoneStat struct {
	Name             string        `json:"zonename"`
	AvgRunQueue1Min  uint32        `json:"avenrun_1min"`    // 1-minute load average
	AvgRunQueue5Min  uint32        `json:"avenrun_5min"`    // 5-minute load average
	AvgRunQueue15Min uint32        `json:"avenrun_15min"`   // 15-minute load average
	BootTime         time.Time     `json:"boot_time"`       // Boot time in seconds since 1970
	CreateTime       float64       `json:"crtime"`          // creation time (from gethrtime())
	ForkFailCap      uint64        `json:"forkfail_cap"`    // hit an rctl cap
	ForkFailMisc     uint64        `json:"forkfail_misc"`   // misc. other error
	ForkFailNoMem    uint64        `json:"forkfail_nomem"`  // as_dup/memory error
	ForkFailNoProc   uint64        `json:"forkfail_noproc"` // get proc/lwp error
	InitPID          uint          `json:"init_pid"`        // PID of "one true" init for current zone
	MapFailSegLimit  uint          `json:"mapfail_seglim"`  // map failure (# segs limit)
	NestedInterp     uint          `json:"nested_interp"`   // nested interp. kstat
	SysTime          time.Duration `json:"nsec_sys"`        //
	UserTime         time.Duration `json:"nsec_user"`       //
	WaitIRQTime      time.Duration `json:"nsec_waitrq"`     //
}

//easyjson:json
type _TimesZoneStat struct {
	TimesZoneStat
	BootTime int64 `json:"boot_time"` // UNIX time
}

//easyjson:json
type _zonesKStatFrame struct {
	Instance uint           `json:"instance"`
	Data     _TimesZoneStat `json:"data"`
}

//easyjson:json
type _zonesKStatFrames []_zonesKStatFrame

// _TimesCPUStatInterrupt has time and counts for 15 levels.  PIL_MAX has been 15
// on Intel and Sparc since the launch of OpenSolaris:
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/intel/sys/machlock.h#L107
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/sparc/sys/machlock.h#L106
//easyjson:json
type _TimesCPUStatInterrupt struct {
	CPUID        uint
	Level1Count  uint64        `json:"level-1-count"`
	Level1Time   time.Duration `json:"level-1-time"`
	Level2Count  uint64        `json:"level-2-count"`
	Level2Time   time.Duration `json:"level-2-time"`
	Level3Count  uint64        `json:"level-3-count"`
	Level3Time   time.Duration `json:"level-3-time"`
	Level4Count  uint64        `json:"level-4-count"`
	Level4Time   time.Duration `json:"level-4-time"`
	Level5Count  uint64        `json:"level-5-count"`
	Level5Time   time.Duration `json:"level-5-time"`
	Level6Count  uint64        `json:"level-6-count"`
	Level6Time   time.Duration `json:"level-6-time"`
	Level7Count  uint64        `json:"level-7-count"`
	Level7Time   time.Duration `json:"level-7-time"`
	Level8Count  uint64        `json:"level-8-count"`
	Level8Time   time.Duration `json:"level-8-time"`
	Level9Count  uint64        `json:"level-9-count"`
	Level9Time   time.Duration `json:"level-9-time"`
	Level10Count uint64        `json:"level-10-count"`
	Level10Time  time.Duration `json:"level-10-time"`
	Level11Count uint64        `json:"level-11-count"`
	Level11Time  time.Duration `json:"level-11-time"`
	Level12Count uint64        `json:"level-12-count"`
	Level12Time  time.Duration `json:"level-12-time"`
	Level13Count uint64        `json:"level-13-count"`
	Level13Time  time.Duration `json:"level-13-time"`
	Level14Count uint64        `json:"level-14-count"`
	Level14Time  time.Duration `json:"level-14-time"`
	Level15Count uint64        `json:"level-15-count"`
	Level15Time  time.Duration `json:"level-15-time"`
}

//easyjson:json
type _cpuInterruptKStatFrame struct {
	Instance uint                   `json:"instance"`
	Data     _TimesCPUStatInterrupt `json:"data"`
}

//easyjson:json
type _cpuInterruptKStatFrames []_cpuInterruptKStatFrame

// TimesCPUStatSys is a struct representing per-CPU stats for sys stats and is
// the output from the "zones" kstat(1M) module.  This list of stats and their
// brief descriptions can be found at:
// https://github.com/illumos/illumos-gate/blob/c65c9cdc54c4ba51aeb34ab36b73286653013c8a/usr/src/uts/common/sys/sysinfo.h#L56-L128
// Some stats also come from usr/src/uts/common/sys/cpuvar.h
//easyjson:json
type TimesCPUStatSys struct {
	CPUID                     uint
	BAWrite                   uint64        `json:"bawrite"`          // physical block writes (async)
	BRead                     uint64        `json:"bread"`            // physical block reads
	BWrite                    uint64        `json:"bwrite"`           // physical block writes (sync+async)
	BytesReadRDRW             uint64        `json:"readch"`           // bytes read by rdwr()
	CAnch                     uint64        `json:"canch"`            // chars handled in canonical mode
	CPULoadInter              uint64        `json:"cpu_load_intr"`    // interrupt load factor (0-99%)
	CPUTMigrate               uint64        `json:"cpumigrate"`       // cpu migrations by threads
	CPUTicksIdle              uint64        `json:"cpu_ticks_idle"`   // Idle CPU utilization
	CPUTicksKernel            uint64        `json:"cpu_ticks_kernel"` // CPU Time in system
	CPUTicksUser              uint64        `json:"cpu_ticks_user"`   // CPU Time in user
	CPUTicksWait              uint64        `json:"cpu_ticks_wait"`   // CPU Time waiting
	CPUTimeDTrace             time.Duration `json:"cpu_nsec_dtrace"`  // DTrace: ns in dtrace_probe
	CPUTimeIdle               time.Duration `json:"cpu_nsec_idle"`    //
	CPUTimeIntr               time.Duration `json:"cpu_nsec_intr"`    //
	CPUTimeKernel             time.Duration `json:"cpu_nsec_kernel"`  //
	CPUTimeUser               time.Duration `json:"cpu_nsec_user"`    //
	CPUWaitIOTicks            uint64        `json:"wait_ticks_io"`    // CPU wait time breakdown
	ContextSwitches           uint64        `json:"pswitch"`          // context switches
	DTraceProbes              uint64        `json:"dtrace_probes"`    // DTrace: total probes fired
	DeviceInterrupts          uint64        `json:"intr"`             // device interrupts
	IOWait                    uint64        `json:"iowait"`           // procs waiting for block I/O
	IdleThread                uint64        `json:"idlethread"`       // times idle thread scheduled
	InterruptBlocked          uint64        `json:"intrblk"`          // ints blkd/prempted/rel'd (swtch)
	InterruptThread           uint64        `json:"intrthread"`       // interrupts as threads (below clock)
	InterruptUnpin            uint64        `json:"intrunpin"`        // intr thread unpins pinned thread
	InvoluntaryCtxtSwitches   uint64        `json:"inv_swtch"`        // involuntary context switches
	LogicalBlockReads         uint64        `json:"lread"`            // logical block reads
	LogicalBlockWrites        uint64        `json:"lwrite"`           // logical block writes
	ModuleLoaded              uint64        `json:"modload"`          // times loadable module loaded
	ModuleUnloaded            uint64        `json:"modunload"`        // times loadable module unloaded
	MsgCount                  uint64        `json:"msg"`              // msg count (msgrcv()+msgsnd() calls)
	MutexFailedAdaptiveEnters uint64        `json:"mutex_adenters"`   // failed mutex enters (adaptive)
	PathnameLookups           uint64        `json:"namei"`            // pathname lookups
	PhysicalReads             uint64        `json:"phread"`           // raw I/O reads
	PhysicalWrites            uint64        `json:"phwrite"`          // raw I/O writes
	ProcTableOverflows        uint64        `json:"procovf"`          // proc table overflows
	RDRWBytes                 uint64        `json:"writech"`          // bytes written by rdwr()
	RWLockReadFails           uint64        `json:"rw_rdfails"`       // rw reader failures
	RWLockWriteFails          uint64        `json:"rw_wrfails"`       // rw writer failures
	SemaphoreOpsCount         uint64        `json:"sema"`             // semaphore ops count (semop() calls)
	SystemCalls               uint64        `json:"syscall"`          // system calls
	SystemExecs               uint64        `json:"sysexec"`          // execs
	SystemForks               uint64        `json:"sysfork"`          // forks
	SystemReads               uint64        `json:"sysread"`          // read() + readv() system calls
	SystemVForks              uint64        `json:"sysvfork"`         // vforks
	SystemWrites              uint64        `json:"syswrite"`         // write() + writev() system calls
	TerminalInputCharacters   uint64        `json:"rawch"`            // terminal input characters
	TerminalOutCharacters     uint64        `json:"outch"`            // terminal output characters
	ThreadsCreated            uint64        `json:"nthreads"`         //  thread_create()s
	Traps                     uint64        `json:"trap"`             // traps
	UFSDirBlocksRead          uint64        `json:"ufsdirblk"`        // directory blocks read
	UFSINodeGets              uint64        `json:"ufsiget"`          // ufs_iget() calls
	UFSINodeNoPages           uint64        `json:"ufsinopage"`       // inodes taked with no attached pages
	UFSINodePage              uint64        `json:"ufsipage"`         // inodes taken with attached pages
	XCalls                    uint64        `json:"xcalls"`           // xcalls to other cpus
}

//easyjson:json
type _cpuSysKStatFrame struct {
	Instance uint            `json:"instance"`
	Data     TimesCPUStatSys `json:"data"`
}

//easyjson:json
type _cpuSysKStatFrames []_cpuSysKStatFrame

// TimesCPUStatVM is a struct representing per-CPU stats for vm stats.  This
// list of stats and their brief descriptions can be found at:
// https://github.com/illumos/illumos-gate/blob/c65c9cdc54c4ba51aeb34ab36b73286653013c8a/usr/src/uts/common/sys/sysinfo.h#L145-L177
//easyjson:json
type TimesCPUStatVM struct {
	CPUID                     uint
	ASMinorPageFault          uint64 `json:"as_fault"`     // minor page faults via as_fault()
	AnonPagesFreed            uint64 `json:"anonfree"`     // anon pages freed
	AnonPagesPagedIn          uint64 `json:"anonpgin"`     // anon pages paged in
	AnonPagesPagedOut         uint64 `json:"anonpgout"`    // anon pages paged out
	COWFaults                 uint64 `json:"cow_fault"`    // copy-on-write faults
	ExecPagesFreed            uint64 `json:"execfree"`     // executable pages freed
	ExecPagesPagedIn          uint64 `json:"execpgin"`     // anon pages paged in
	ExecPagesPagedOut         uint64 `json:"execpgout"`    // executable pages paged out
	FSPagesFreed              uint64 `json:"fsfree"`       // fs pages free
	FSPagesPagedIn            uint64 `json:"fspgin"`       // fs pages paged in
	FSPagesPagedOut           uint64 `json:"fspgout"`      // fs pages paged out
	HATMinorPageFaults        uint64 `json:"hat_fault"`    // minor page faults via hat_fault()
	KernelASFaults            uint64 `json:"kernel_asflt"` // as_fault()s in kernel addr space
	MajorPageFaults           uint64 `json:"maj_fault"`    // major page faults
	PageDaemonHandRevolutions uint64 `json:"rev"`          // revolutions of the page daemon hand
	PageDaemonPagesScanned    uint64 `json:"scan"`         // pages examined by pageout daemon
	PageFreeListReclaims      uint64 `json:"pgfrec"`       // page reclaims from free list
	PageIns                   uint64 `json:"pgin"`         // pageins
	PageOuts                  uint64 `json:"pgout"`        // pageouts
	PagerScheduled            uint64 `json:"pgrrun"`       // times pager scheduled
	PagesFreed                uint64 `json:"dfree"`        // pages freed by daemon or auto
	PagesPagedIn              uint64 `json:"pgpgin"`       // pages paged in
	PagesPagedOut             uint64 `json:"pgpgout"`      // pages paged out
	PagesReclaimed            uint64 `json:"pgrec"`        // page reclaims (includes pageout)
	PagesSwappedIn            uint64 `json:"pgswapin"`     // pages swapped in
	PagesSwappedOut           uint64 `json:"pgswapout"`    // pages swapped out
	PagesZeroFilledOnDemand   uint64 `json:"zfod"`         // pages zero filled on demand
	ProtectionFaults          uint64 `json:"prot_fault"`   // protection faults
	SoftwareLockFaults        uint64 `json:"softlock"`     // faults due to software locking req
	SwapIns                   uint64 `json:"swapin"`       // swapins
	SwapOuts                  uint64 `json:"swapout"`      // swapouts
}

//easyjson:json
type _cpuVMKStatFrame struct {
	Instance uint           `json:"instance"`
	Data     TimesCPUStatVM `json:"data"`
}

//easyjson:json
type _cpuVMKStatFrames []_cpuVMKStatFrame

//easyjson:json
type TimesCPUStat struct {
	CPUID uint
	TimesCPUStatInterrupt
	TimesCPUStatSys
	TimesCPUStatVM
}
