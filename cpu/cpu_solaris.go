package cpu

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

type TimesZoneStat struct {
	Name             string        `json:"zonename"`
	AvgRunQueue1Min  int32         `json:"avenrun_1min"`    // 1-minute load average
	AvgRunQueue5Min  int32         `json:"avenrun_5min"`    // 5-minute load average
	AvgRunQueue15Min int32         `json:"avenrun_15min"`   // 15-minute load average
	BootTime         time.Time     `json:"boot_time"`       // Boot time in seconds since 1970
	CreateTime       float64       `json:"crtime"`          // creation time (from gethrtime())
	ForkFailCap      int64         `json:"forkfail_cap"`    // hit an rctl cap
	ForkFailMisc     int64         `josn:"forkfail_misc"`   // misc. other error
	ForkFailNoMem    int64         `json:"forkfail_nomem"`  // as_dup/memory error
	ForkFailNoProc   int64         `json:"forkfail_noproc"` // get proc/lwp error
	InitPID          int           `json:"init_pid"`        // PID of "one true" init for current zone
	MapFailSegLimit  int           `json:"mapfail_seglim"`  // map failure (# segs limit)
	NestedInterp     int           `json:"nested_interp"`   // nested interp. kstat
	SysTime          time.Duration `json:"nsec_sys"`        //
	UserTime         time.Duration `json:"nsec_user"`       //
	WaitIRQTime      time.Duration `json:"nsec_waitrq"`     //
}

// PIL_MAX has been 15 on Intel and Sparc since the launch of OpenSolaris:
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/intel/sys/machlock.h#L107
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/sparc/sys/machlock.h#L106
const PIL_MAX = 15

type TimesCPUStatInterruptLevel struct {
	Count uint64        // Number of times an interrupt at a given level was triggered
	Time  time.Duration // Time spent in interrupt context at this level
}

// TimesCPUStatInterrupt has Time and Counts for 15 Processor Interrupt Levels,
// zero-based indexing.
type TimesCPUStatInterrupt struct {
	CPUID int
	PIL   []TimesCPUStatInterruptLevel // Processor Interrupt Level
}

// _TimesCPUStatInterrupt has time and counts for 15 levels.  PIL_MAX has been 15
// on Intel and Sparc since the launch of OpenSolaris:
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/intel/sys/machlock.h#L107
// https://github.com/illumos/illumos-gate/blame/2428aad8462660fad2b105777063fea6f4192308/usr/src/uts/sparc/sys/machlock.h#L106
type _TimesCPUStatInterrupt struct {
	CPUID        int
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

// TimesCPUStatSys is a struct representing per-CPU stats for sys stats.  This
// list of stats and their brief descriptions can be found at:
// https://github.com/illumos/illumos-gate/blob/c65c9cdc54c4ba51aeb34ab36b73286653013c8a/usr/src/uts/common/sys/sysinfo.h#L56-L128
// Some stats also come from usr/src/uts/common/sys/cpuvar.h
type TimesCPUStatSys struct {
	CPUID                     int
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
	ProcTableOverflows        uint64        `josn:"procovf"`          // proc table overflows
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

// TimesCPUStatVM is a struct representing per-CPU stats for vm stats.  This
// list of stats and their brief descriptions can be found at:
// https://github.com/illumos/illumos-gate/blob/c65c9cdc54c4ba51aeb34ab36b73286653013c8a/usr/src/uts/common/sys/sysinfo.h#L145-L177
type TimesCPUStatVM struct {
	CPUID                     int
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

type TimesCPUStat struct {
	CPUID int
	TimesCPUStatInterrupt
	TimesCPUStatSys
	TimesCPUStatVM
}

var (
	// Path to various utilities
	isainfoPath string
	kstatPath   string
	psrinfoPath string

	// inZone is true when it has been detected that a process is running inside
	// of an Illumos zone (see zones(5)).
	inZone bool

	// Only stat(2) the locations of binaries once.  All exported interfaces that
	// shell out to a binary must call through and check that that paths have been
	// resolved.
	lookupOnce sync.Once
	pathErr    error // Shared error from any exec.LookPath() call
)

func init() {
	// ignore errors
	switch true {
	default:
		zonenamePath, err := exec.LookPath("/usr/bin/zonename")
		if err != nil {
			break
		}

		out, err := invoke.Command(zonenamePath)
		if err != nil {
			break
		}

		if string(out) != "global" {
			inZone = true
		}
	}
}

func lookupPaths() {
	psrinfoPath, pathErr = exec.LookPath("/usr/sbin/psrinfo")
	if pathErr != nil {
		return
	}

	isainfoPath, pathErr = exec.LookPath("/usr/bin/isainfo")
	if pathErr != nil {
		return
	}

	kstatPath, pathErr = exec.LookPath("/usr/bin/kstat")
	if pathErr != nil {
		return
	}
}

func Times(perCPU bool) ([]TimesStat, error) {
	var ret []TimesStat
	var err error

	switch {
	case perCPU:
		var cpuStats []TimesCPUStatSys
		cpuStats, err = TimesPerCPUSys()
		if err != nil {
			return nil, err
		}
		ret = make([]TimesStat, 0, len(cpuStats))

		for _, cpuStat := range cpuStats {
			ret = append(ret, TimesStat{
				CPU:       fmt.Sprintf("cpu%d", cpuStat.CPUID),
				User:      float64(time.Duration(cpuStat.CPUTimeUser) / time.Second),
				System:    float64(time.Duration(cpuStat.CPUTimeKernel) / time.Second),
				Idle:      float64(time.Duration(cpuStat.CPUTimeIdle) / time.Second),
				Nice:      0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Iowait:    float64(time.Duration(cpuStat.IOWait) / time.Second),
				Irq:       float64(time.Duration(cpuStat.DeviceInterrupts) / time.Second),
				Softirq:   float64(time.Duration(cpuStat.CPUTimeIntr) / time.Second),
				Steal:     0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Guest:     0.0, // NOTE(seanc@): no value emitted by Illumos at present
				GuestNice: 0.0, // NOTE(seanc@): no value emitted by Illumos at present
				Stolen:    0.0, // NOTE(seanc@): no value emitted by Illumos at present
			})
		}
	case !perCPU && inZone:
		var zoneStats []TimesZoneStat
		zoneStats, err = TimesZone()
		if err != nil {
			return nil, err
		}

		ret = make([]TimesStat, 0, len(zoneStats))

		for i, zone := range zoneStats {
			ret = append(ret, TimesStat{
				CPU:    fmt.Sprintf("zone%d", i),
				User:   float64(zone.UserTime / time.Second),
				System: float64(zone.SysTime / time.Second),
				Irq:    float64(zone.WaitIRQTime / time.Second),
			})
		}
	default:
		var cpuStats []TimesCPUStat
		cpuStats, err = TimesCPU()
		if err != nil {
			return nil, err
		}
		timeStat := TimesStat{}

		timeStat.CPU = "cpuN"
		for _, cpuStat := range cpuStats {
			timeStat.User += float64(time.Duration(cpuStat.CPUTimeUser) / time.Second)
			timeStat.System += float64(time.Duration(cpuStat.CPUTimeKernel) / time.Second)
			timeStat.Irq += float64(time.Duration(cpuStat.CPUTimeIntr) / time.Second)
		}

		ret = []TimesStat{timeStat}
	}

	return ret, nil
}

// TimesPerCPU returns the CPU time spent for each processor.  This information
// is extremely granular (and useful!).
func TimesPerCPU() ([]TimesCPUStat, error) {
	lookupOnce.Do(lookupPaths)
	if len(kstatPath) == 0 {
		return nil, fmt.Errorf("cannot find kstat: %v", pathErr)
	}

	// Make three kstat(1) calls: intrstat, sys, vm
	interruptStats, err := TimesPerCPUInterrupt()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU interrupt stats: %v", err)
	}

	sysStats, err := TimesPerCPUSys()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU system stats: %v", err)
	}

	vmStats, err := TimesPerCPUVM()
	if err != nil {
		return nil, fmt.Errorf("cannot get CPU VM stats: %v", err)
	}

	if len(interruptStats) != len(sysStats) && len(sysStats) != len(vmStats) {
		return nil, fmt.Errorf("length of stats not identical: invariant broken")
	}

	ret := make([]TimesCPUStat, len(interruptStats))
	for i := 0; i < len(ret); i++ {
		ret[i] = TimesCPUStat{
			CPUID: i,
			TimesCPUStatInterrupt: interruptStats[i],
			TimesCPUStatSys:       sysStats[i],
			TimesCPUStatVM:        vmStats[i],
		}
	}

	return ret, nil
}

func TimesPerCPUInterrupt() ([]TimesCPUStatInterrupt, error) {
	lookupOnce.Do(lookupPaths)
	if len(kstatPath) == 0 {
		return nil, fmt.Errorf("cannot find kstat: %s", pathErr)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "intrstat")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	type _kstatFrame struct {
		Instance int                    `json:"instance"`
		Data     _TimesCPUStatInterrupt `json:"data"`
	}
	kstatFrames := make([]_kstatFrame, 0, 1)
	err = json.Unmarshal(kstatJSON, &kstatFrames)
	if err != nil {
		return nil, err
	}

	ret := make([]TimesCPUStatInterrupt, len(kstatFrames))
	for cpuNum, pilStat := range kstatFrames {
		if pilStat.Instance != cpuNum {
			return nil, fmt.Errorf("kstat CPU intrstat output not ordered by instance: invariant broken")
		}

		ret[cpuNum].CPUID = cpuNum

		pil := make([]TimesCPUStatInterruptLevel, PIL_MAX)
		psd := pilStat.Data
		pil[0].Count = psd.Level1Count
		pil[0].Time = psd.Level1Time
		pil[1].Count = psd.Level2Count
		pil[1].Time = psd.Level2Time
		pil[2].Count = psd.Level3Count
		pil[2].Time = psd.Level3Time
		pil[3].Count = psd.Level4Count
		pil[3].Time = psd.Level4Time
		pil[4].Count = psd.Level5Count
		pil[4].Time = psd.Level5Time
		pil[5].Count = psd.Level6Count
		pil[5].Time = psd.Level6Time
		pil[6].Count = psd.Level7Count
		pil[6].Time = psd.Level7Time
		pil[7].Count = psd.Level8Count
		pil[7].Time = psd.Level8Time
		pil[8].Count = psd.Level9Count
		pil[8].Time = psd.Level9Time
		pil[9].Count = psd.Level10Count
		pil[9].Time = psd.Level10Time
		pil[10].Count = psd.Level11Count
		pil[10].Time = psd.Level11Time
		pil[11].Count = psd.Level12Count
		pil[11].Time = psd.Level12Time
		pil[12].Count = psd.Level13Count
		pil[12].Time = psd.Level13Time
		pil[13].Count = psd.Level14Count
		pil[13].Time = psd.Level14Time
		pil[14].Count = psd.Level15Count
		pil[14].Time = psd.Level15Time

		ret[cpuNum].PIL = pil
	}

	return ret, nil
}

func TimesPerCPUSys() ([]TimesCPUStatSys, error) {
	lookupOnce.Do(lookupPaths)
	if len(kstatPath) == 0 {
		return nil, fmt.Errorf("cannot find kstat: %s", pathErr)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "sys")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	type _kstatFrame struct {
		Instance int             `json:"instance"`
		Data     TimesCPUStatSys `json:"data"`
	}
	kstatFrames := make([]_kstatFrame, 0, 1)
	err = json.Unmarshal(kstatJSON, &kstatFrames)
	if err != nil {
		return nil, err
	}

	ret := make([]TimesCPUStatSys, 0, len(kstatFrames))
	for cpuNum, sysStat := range kstatFrames {
		if sysStat.Instance != cpuNum {
			return nil, fmt.Errorf("kstat CPU sys output not ordered by instance: invariant broken")
		}
		sysStat.Data.CPUID = cpuNum
		ret = append(ret, sysStat.Data)
	}

	return ret, nil
}

func TimesPerCPUVM() ([]TimesCPUStatVM, error) {
	lookupOnce.Do(lookupPaths)
	if len(kstatPath) == 0 {
		return nil, fmt.Errorf("cannot find kstat: %s", pathErr)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "cpu", "-n", "vm")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	type _kstatFrame struct {
		Instance int            `json:"instance"`
		Data     TimesCPUStatVM `json:"data"`
	}
	kstatFrames := make([]_kstatFrame, 0, 1)
	err = json.Unmarshal(kstatJSON, &kstatFrames)
	if err != nil {
		return nil, err
	}

	ret := make([]TimesCPUStatVM, 0, len(kstatFrames))
	for cpuNum, vmStat := range kstatFrames {
		if vmStat.Instance != cpuNum {
			return nil, fmt.Errorf("kstat CPU vm output not ordered by instance: invariant broken")
		}
		vmStat.Data.CPUID = cpuNum
		ret = append(ret, vmStat.Data)
	}

	return ret, nil
}

// TimesCPU returns the CPU time spent on all processors.  This information is
// extremely granular (and useful!).
func TimesCPU() ([]TimesCPUStat, error) {
	lookupOnce.Do(lookupPaths)
	allTimes, err := TimesPerCPU()
	if err != nil {
		return nil, err
	}

	return allTimes, nil
}

// TimesZone returns the CPU time spent in the visible zone(s).  In a zone we
// can't get per-CPU accounting.  Instead, report back an accurate use of CPU
// time for the given zone(s) but only report one CPU, the "zoneN" CPU.
func TimesZone() ([]TimesZoneStat, error) {
	lookupOnce.Do(lookupPaths)
	if len(kstatPath) == 0 {
		return nil, fmt.Errorf("cannot find kstat: %s", pathErr)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "zones")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat: %v", err)
	}

	type _timesZoneStat struct {
		TimesZoneStat
		BootTime int64 `json:"boot_time"`
	}
	type _kstatFrame struct {
		Instance int            `json:"instance"`
		Data     _timesZoneStat `json:"data"`
	}
	kstatFrames := make([]_kstatFrame, 0, 1)
	err = json.Unmarshal(kstatJSON, &kstatFrames)
	if err != nil {
		return nil, err
	}

	ret := make([]TimesZoneStat, 0, len(kstatFrames))
	for _, zoneStat := range kstatFrames {
		zStat := zoneStat.Data.TimesZoneStat
		zStat.BootTime = time.Unix(zoneStat.Data.BootTime, 0)
		ret = append(ret, zStat)
	}

	return ret, nil
}

func Info() ([]InfoStat, error) {
	lookupOnce.Do(lookupPaths)
	if len(psrinfoPath) == 0 {
		return nil, fmt.Errorf("cannot find psrinfo: %s", pathErr)
	}
	psrinfoOut, err := invoke.Command(psrinfoPath, "-p", "-v")
	if err != nil {
		return nil, fmt.Errorf("cannot execute psrinfo: %s", err)
	}

	if len(isainfoPath) == 0 {
		return nil, fmt.Errorf("cannot find isainfo: %s", pathErr)
	}
	isainfoOut, err := invoke.Command(isainfoPath, "-b", "-v")
	if err != nil {
		return nil, fmt.Errorf("cannot execute isainfo: %s", err)
	}

	procs, err := parseProcessorInfo(string(psrinfoOut))
	if err != nil {
		return nil, fmt.Errorf("error parsing psrinfo output: %s", err)
	}

	flags, err := parseISAInfo(string(isainfoOut))
	if err != nil {
		return nil, fmt.Errorf("error parsing isainfo output: %s", err)
	}

	result := make([]InfoStat, 0, len(flags))
	for _, proc := range procs {
		procWithFlags := proc
		procWithFlags.Flags = flags
		result = append(result, procWithFlags)
	}

	return result, nil
}

var flagsMatch = regexp.MustCompile(`[\w\.]+`)

func parseISAInfo(cmdOutput string) ([]string, error) {
	words := flagsMatch.FindAllString(cmdOutput, -1)

	// Sanity check the output
	if len(words) < 4 || words[1] != "bit" || words[3] != "applications" {
		return nil, errors.New("attempted to parse invalid isainfo output")
	}

	flags := make([]string, len(words)-4)
	for i, val := range words[4:] {
		flags[i] = val
	}
	sort.Strings(flags)

	return flags, nil
}

var psrinfoMatch = regexp.MustCompile(`The physical processor has (?:([\d]+) virtual processor \(([\d]+)\)|([\d]+) cores and ([\d]+) virtual processors[^\n]+)\n(?:\s+ The core has.+\n)*\s+.+ \((\w+) ([\S]+) family (.+) model (.+) step (.+) clock (.+) MHz\)\n[\s]*(.*)`)

const (
	psrNumCoresOffset   = 1
	psrNumCoresHTOffset = 3
	psrNumHTOffset      = 4
	psrVendorIDOffset   = 5
	psrFamilyOffset     = 7
	psrModelOffset      = 8
	psrStepOffset       = 9
	psrClockOffset      = 10
	psrModelNameOffset  = 11
)

func parseProcessorInfo(cmdOutput string) ([]InfoStat, error) {
	matches := psrinfoMatch.FindAllStringSubmatch(cmdOutput, -1)

	var infoStatCount int32
	result := make([]InfoStat, 0, len(matches))
	for physicalIndex, physicalCPU := range matches {
		var step int32
		var clock float64

		if physicalCPU[psrStepOffset] != "" {
			stepParsed, err := strconv.ParseInt(physicalCPU[psrStepOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for step as 32-bit integer: %s", physicalCPU[9], err)
			}
			step = int32(stepParsed)
		}

		if physicalCPU[psrClockOffset] != "" {
			clockParsed, err := strconv.ParseInt(physicalCPU[psrClockOffset], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for clock as 32-bit integer: %s", physicalCPU[10], err)
			}
			clock = float64(clockParsed)
		}

		var err error
		var numCores int64
		var numHT int64
		switch {
		case physicalCPU[psrNumCoresOffset] != "":
			numCores, err = strconv.ParseInt(physicalCPU[psrNumCoresOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for core count as 32-bit integer: %s", physicalCPU[1], err)
			}

			for i := 0; i < int(numCores); i++ {
				result = append(result, InfoStat{
					CPU:        infoStatCount,
					PhysicalID: strconv.Itoa(physicalIndex),
					CoreID:     strconv.Itoa(i),
					Cores:      1,
					VendorID:   physicalCPU[psrVendorIDOffset],
					ModelName:  physicalCPU[psrModelNameOffset],
					Family:     physicalCPU[psrFamilyOffset],
					Model:      physicalCPU[psrModelOffset],
					Stepping:   step,
					Mhz:        clock,
				})
				infoStatCount++
			}
		case physicalCPU[psrNumCoresHTOffset] != "":
			numCores, err = strconv.ParseInt(physicalCPU[psrNumCoresHTOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for core count as 32-bit integer: %s", physicalCPU[3], err)
			}

			numHT, err = strconv.ParseInt(physicalCPU[psrNumHTOffset], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("cannot parse value %q for hyperthread count as 32-bit integer: %s", physicalCPU[4], err)
			}

			for i := 0; i < int(numCores); i++ {
				result = append(result, InfoStat{
					CPU:        infoStatCount,
					PhysicalID: strconv.Itoa(physicalIndex),
					CoreID:     strconv.Itoa(i),
					Cores:      int32(numHT) / int32(numCores),
					VendorID:   physicalCPU[psrVendorIDOffset],
					ModelName:  physicalCPU[psrModelNameOffset],
					Family:     physicalCPU[psrFamilyOffset],
					Model:      physicalCPU[psrModelOffset],
					Stepping:   step,
					Mhz:        clock,
				})
				infoStatCount++
			}
		default:
			return nil, errors.New("values for cores with and without hyperthreading are both set")
		}
	}
	return result, nil
}
