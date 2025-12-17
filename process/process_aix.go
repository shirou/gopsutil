// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/internal/common"
	"github.com/shirou/gopsutil/v4/net"
)

var pageSize = uint64(os.Getpagesize())

// AIX-specific: cache for process bitness (4 = 32-bit, 8 = 64-bit)
var aixBitnessCache sync.Map // map[int32]int64

const prioProcess = 0 // linux/resource.h

var clockTicks = 100 // default value

func init() {
	// Initialize clock ticks from AIX schedo configuration
	// AIX default: 1 clock tick = 10ms (100 ticks/second)
	// Can be modified via schedo big_tick_size parameter
	clockTicks = getAIXClockTicks()
}

// getAIXClockTicks retrieves the actual clock tick frequency from AIX scheduler configuration.
// AIX maintains this through the schedo command, specifically the big_tick_size parameter.
// The default is 1 tick = 10ms, but this can be tuned via:
//
//	schedo -o big_tick_size=<value>
//
// where value * 10ms is the actual tick interval.
//
// Since we cannot directly access kernel parameters from userspace reliably,
// we use the schedo command to query big_tick_size and calculate the actual clock ticks.
// If schedo is unavailable or returns an error, we default to the standard 100 ticks/second (10ms).
func getAIXClockTicks() int {
	const defaultClockTicks = 100 // Default: 100 ticks/second = 10ms per tick

	// Try to query big_tick_size from schedo
	// Format: schedo -o big_tick_size (displays current value without changing)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "schedo", "-o", "big_tick_size")

	output, err := cmd.Output()
	if err != nil {
		// schedo unavailable or failed; use default
		return defaultClockTicks
	}

	// Parse output format: "big_tick_size = <value>"
	// Example: "big_tick_size = 1"
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Split(outputStr, "=")
	if len(parts) < 2 {
		return defaultClockTicks
	}

	tickMultiplier, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return defaultClockTicks
	}

	// big_tick_size is a multiplier for 10ms ticks
	// Calculate actual clock ticks per second: 1000ms / (tickMultiplier * 10ms)
	if tickMultiplier <= 0 {
		return defaultClockTicks
	}

	// 100 ticks/second / tickMultiplier = actual ticks per second
	// For example: if big_tick_size=2, then 100/2 = 50 ticks/second
	actualTicks := defaultClockTicks / tickMultiplier
	if actualTicks < 1 {
		actualTicks = 1 // Ensure at least 1 tick per second
	}

	return actualTicks
}

type PrTimestruc64T struct {
	TvSec  int64  // 64 bit time_t value
	TvNsec int32  // 32 bit suseconds_t value
	Pad    uint32 // reserved for future use
}

/* hardware fault set */
type Fltset struct {
	FltSet [4]uint64 // fault set
}

type PrSigset struct {
	SsSet [4]uint64 // signal set
}

type prptr64 uint64

type size64 uint64

type pid64 uint64 // size invariant 64-bit pid

type PrSiginfo64 struct {
	SiSigno  int32     // signal number
	SiErrno  int32     // if non-zero, errno of this signal
	SiCode   int32     // signal code
	SiImm    int32     // immediate data
	SiStatus int32     // exit value or signal
	Pad1     uint32    // reserved for future use
	SiUID    uint64    // real user id of sending process
	SiPid    uint64    // sending process id
	SiAddr   prptr64   // address of faulting instruction
	SiBand   int64     // band event for SIGPOLL
	SiValue  prptr64   // signal value
	Pad      [4]uint32 // reserved for future use
}

type PrSigaction64 struct {
	SaUnion prptr64  // signal handler function pointer
	SaMask  PrSigset // signal mask
	SaFlags int32    // signal flags
	Pad     [5]int32 // reserved for future use
}

type PrStack64 struct {
	SsSp    prptr64  // stack base or pointer
	SsSize  uint64   // stack size
	SsFlags int32    // flags
	Pad     [5]int32 // reserved for future use
}

type Prgregset struct {
	Iar    size64     // Instruction Pointer
	Msr    size64     // machine state register
	Cr     size64     // condition register
	Lr     size64     // link register
	Ctr    size64     // count register
	Xer    size64     // fixed point exception
	Fpscr  size64     // floating point status reg
	Fpscrx size64     // extension floating point
	Gpr    [32]size64 // static general registers
	Usprg3 size64
	Pad1   [7]size64 // Reserved for future use
}

type Prfpregset struct {
	Fpr [32]float64 // Floating Point Registers
}

type Pfamily struct {
	ExtOff  uint64     // offset of extension
	ExtSize uint64     // size of extension
	Pad     [14]uint64 // reserved for future use
}

type LwpStatus struct {
	LwpId    uint64        // specific thread id
	Flags    uint32        // thread status flags
	Pad1     [1]byte       // reserved for future use
	State    byte          // thread state
	CurSig   uint16        // current signal
	Why      uint16        // stop reason
	What     uint16        // more detailed reason
	Policy   uint32        // scheduling policy
	ClName   [8]byte       // scheduling policy string
	LwpPend  PrSigset      // set of signals pending to the thread
	LwpHold  PrSigset      // set of signals blocked by the thread
	Info     PrSiginfo64   // info associated with signal or fault
	AltStack PrStack64     // alternate signal stack info
	Action   PrSigaction64 // signal action for current signal
	Pad2     uint32        // reserved for future use
	Syscall  uint16        // system call number
	NsysArg  uint16        // number of arguments
	SysArg   [8]uint64     // syscall arguments
	Errno    int32         // errno from last syscall
	Ptid     uint32        // pthread id
	Pad      [9]uint64     // reserved for future use
	Reg      Prgregset     // general registers
	Fpreg    Prfpregset    // floating point registers
	Family   Pfamily       // hardware platform specific information
}

type AIXStat struct {
	Flag           uint32         // process flags from proc struct p_flag
	Flag2          uint32         // process flags from proc struct p_flag2
	Flags          uint32         // /proc flags
	Nlwp           uint32         // number of threads in the process
	Stat           byte           // process state from proc p_stat
	Dmodel         byte           // data model for the process
	Pad1           [6]byte        // reserved for future use
	SigPend        PrSigset       // set of process pending signals
	BrkBase        prptr64        // address of the process heap
	BrkSize        uint64         // size of the process heap, in bytes
	StkBase        prptr64        // address of the process stack
	StkSize        uint64         // size of the process stack, in bytes
	Pid            pid64          // process id
	Ppid           pid64          // parent process id
	Pgid           pid64          // process group id
	Sid            pid64          // session id
	Utime          PrTimestruc64T // process user cpu time
	Stime          PrTimestruc64T // process system cpu time
	Cutime         PrTimestruc64T // sum of children's user times
	Cstime         PrTimestruc64T // sum of children's system times
	SigTrace       PrSigset       // mask of traced signals
	FltTrace       Fltset         // mask of traced hardware faults
	SysentryOffset uint32         // offset into pstatus file of sysset_t identifying system calls traced on entry
	SysexitOffset  uint32         // offset into pstatus file of sysset_t identifying system calls traced on exit
	Pad            [8]uint64      // reserved for future use
	Lwp            LwpStatus      // "representative" thread status
}

type LwpsInfo struct {
	LwpId   uint64    // thread id
	Addr    uint64    // internal address of thread
	Wchan   uint64    // wait address for sleeping thread
	Flag    uint32    // thread flags
	Wtype   byte      // type of thread wait
	State   byte      // thread state
	Sname   byte      // printable thread state character
	Nice    byte      // nice value for CPU usage
	Pri     int32     // priority, high value = high priority
	Policy  uint32    // scheduling policy
	Clname  [8]byte   // printable scheduling policy string
	Onpro   int32     // processor on which thread last ran
	Bindpro int32     // processor to which thread is bound
	Ptid    uint32    // pthread id
	Pad1    uint32    // reserved for future use
	Pad     [7]uint64 // reserved for future use
}

type AIXPSInfo struct {
	Flag   uint32         // process flags from proc struct p_flag
	Flag2  uint32         // process flags from proc struct p_flag2
	Nlwp   uint32         // number of threads in process
	Pad1   uint32         // reserved for future use
	UID    uint64         // real user id
	Euid   uint64         // effective user id
	Gid    uint64         // real group id
	Egid   uint64         // effective group id
	Pid    uint64         // unique process id
	Ppid   uint64         // process id of parent
	Pgid   uint64         // pid of process group leader
	Sid    uint64         // session id
	Ttydev uint64         // controlling tty device
	Addr   uint64         // internal address of proc struct
	Size   uint64         // size of process image in KB (1024) units
	Rssize uint64         // resident set size in KB (1024) units
	Start  PrTimestruc64T // process start time, time since epoch
	Time   PrTimestruc64T // usr+sys cpu time for this process
	Cid    int16          // corral id
	Pad2   int16          // reserved for future use
	Argc   uint32         // initial argument count
	Argv   uint64         // address of initial argument vector in user process
	Envp   uint64         // address of initial environment vector in user process
	Fname  [16]byte       // last component of exec()ed pathname
	Psargs [80]byte       // initial characters of arg list
	Pad    [8]uint64      // reserved for future use
	Lwp    LwpsInfo       // "representative" thread info
}

// MemoryInfoExStat is different between OSes
type MemoryInfoExStat struct {
	RSS    uint64 `json:"rss"`    // bytes
	VMS    uint64 `json:"vms"`    // bytes
	Shared uint64 `json:"shared"` // bytes
	Text   uint64 `json:"text"`   // bytes
	Lib    uint64 `json:"lib"`    // bytes
	Data   uint64 `json:"data"`   // bytes
	Dirty  uint64 `json:"dirty"`  // bytes
}

func (m MemoryInfoExStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

type MemoryMapsStat struct {
	Path         string `json:"path"`
	Rss          uint64 `json:"rss"`
	Size         uint64 `json:"size"`
	Pss          uint64 `json:"pss"`
	SharedClean  uint64 `json:"sharedClean"`
	SharedDirty  uint64 `json:"sharedDirty"`
	PrivateClean uint64 `json:"privateClean"`
	PrivateDirty uint64 `json:"privateDirty"`
	Referenced   uint64 `json:"referenced"`
	Anonymous    uint64 `json:"anonymous"`
	Swap         uint64 `json:"swap"`
}

// String returns JSON value of the process.
func (m MemoryMapsStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (p *Process) PpidWithContext(ctx context.Context) (int32, error) {
	_, ppid, _, _, _, _, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return -1, err
	}
	return ppid, nil
}

func (p *Process) NameWithContext(ctx context.Context) (string, error) {
	if p.name == "" {
		if err := p.fillFromCommWithContext(ctx); err != nil {
			return "", err
		}
	}
	return p.name, nil
}

func (p *Process) TgidWithContext(ctx context.Context) (int32, error) {
	if p.tgid == 0 {
		if err := p.fillFromStatusWithContext(ctx); err != nil {
			return 0, err
		}
	}
	return p.tgid, nil
}

func (p *Process) ExeWithContext(ctx context.Context) (string, error) {
	return p.fillFromExeWithContext(ctx)
}

func (p *Process) CmdlineWithContext(ctx context.Context) (string, error) {
	return p.fillFromCmdlineWithContext(ctx)
}

func (p *Process) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	return p.fillSliceFromCmdlineWithContext(ctx)
}

func (p *Process) EnvironmentWithContext(ctx context.Context) (map[string]string, error) {
	// Query environment via ps command using Berkeley-style 'e' option
	// Berkeley style: ps eww <PID> (no -p flag)
	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd := exec.CommandContext(ctx, "ps", "eww", strconv.Itoa(int(p.Pid)))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	envStr := strings.TrimSpace(string(output))
	if envStr == "" {
		return make(map[string]string), nil
	}

	// Parse space-separated VAR=value assignments
	env := make(map[string]string)

	// ps eww output is space-separated on a single line (or multiple lines for multiline values)
	// Split by spaces to get individual VAR=value pairs
	parts := strings.Fields(envStr)

	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				env[kv[0]] = kv[1]
			}
		}
	}

	return env, nil
}

func (p *Process) createTimeWithContext(ctx context.Context) (int64, error) {
	_, _, _, createTime, _, _, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return createTime, nil
}

func (p *Process) CwdWithContext(ctx context.Context) (string, error) {
	return p.fillFromCwdWithContext(ctx)
}

func (p *Process) StatusWithContext(ctx context.Context) ([]string, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return []string{""}, err
	}
	return []string{p.status}, nil
}

func (*Process) ForegroundWithContext(_ context.Context) (bool, error) {
	return false, common.ErrNotImplementedError
}

func (p *Process) UidsWithContext(ctx context.Context) ([]uint32, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return []uint32{}, err
	}
	return p.uids, nil
}

func (p *Process) GidsWithContext(ctx context.Context) ([]uint32, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return []uint32{}, err
	}
	return p.gids, nil
}

func (p *Process) GroupsWithContext(ctx context.Context) ([]uint32, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return []uint32{}, err
	}
	return p.groups, nil
}

func (p *Process) TerminalWithContext(ctx context.Context) (string, error) {
	// Query TTY via ps command
	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd := exec.CommandContext(ctx, "ps", "-o", "tty", "-p", strconv.Itoa(int(p.Pid)))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		// Only header or no output
		return "", nil
	}

	// Get the TTY value (second line, first field)
	tty := strings.Fields(lines[1])
	if len(tty) > 0 {
		return tty[0], nil
	}
	return "", nil
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	_, _, _, _, _, nice, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return nice, nil
}

func (*Process) IOniceWithContext(_ context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	return p.RlimitUsageWithContext(ctx, false)
}

func (p *Process) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	// Get per-process resource limits via procfiles command
	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd := exec.CommandContext(ctx, "procfiles", strconv.Itoa(int(p.Pid)))
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try system-wide limits from ulimit
		return p.getRlimitFromUlimit(ctx, gatherUsed)
	}

	// Parse procfiles output for file descriptor limits
	// Output format: FD Info: nnnn (soft limit), mmmm (hard limit)
	var rlimits []RlimitStat

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.Contains(line, "FD Info") || strings.Contains(line, "File descriptors") {
			// Extract limits from this line
			// Format varies, try to parse numbers
			numStrs := strings.FieldsFunc(line, func(r rune) bool {
				return r < '0' || r > '9'
			})
			if len(numStrs) >= 2 {
				soft, _ := strconv.ParseUint(numStrs[0], 10, 64)
				hard, _ := strconv.ParseUint(numStrs[1], 10, 64)
				rlimits = append(rlimits, RlimitStat{
					Resource: RLIMIT_NOFILE,
					Soft:     soft,
					Hard:     hard,
				})
				break
			}
		}
	}

	if len(rlimits) == 0 {
		return p.getRlimitFromUlimit(ctx, gatherUsed)
	}

	return rlimits, nil
}

// getRlimitFromUlimit gets resource limits via ulimit command
func (*Process) getRlimitFromUlimit(ctx context.Context, _ bool) ([]RlimitStat, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", "ulimit -a")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var rlimits []RlimitStat
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		// Parse ulimit output: "open files (-n) 1024"
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		var limit uint64

		// Identify resource type and extract limit
		var resourceID int32
		if strings.Contains(line, "open files") {
			resourceID = RLIMIT_NOFILE
			// Extract numeric value (usually last field)
			if val, err := strconv.ParseUint(fields[len(fields)-1], 10, 64); err == nil {
				limit = val
			}
		}

		if resourceID != 0 && limit > 0 {
			rlimits = append(rlimits, RlimitStat{
				Resource: resourceID,
				Soft:     limit,
				Hard:     limit,
			})
		}
	}

	return rlimits, nil
}

func (p *Process) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	// Check if WLM is enabled and iostat is configured
	cmd := exec.CommandContext(ctx, "lsattr", "-El", "sys0")
	output, err := cmd.Output()
	if err != nil {
		return nil, common.ErrNotImplementedError
	}

	// Check if iostat=true
	if !strings.Contains(string(output), "iostat true") {
		return nil, common.ErrNotImplementedError
	}

	// Query I/O counters via ps command
	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd = exec.CommandContext(ctx, "ps", "-efo", "pid,tdiskio", "-p", strconv.Itoa(int(p.Pid)))
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, errors.New("insufficient ps output")
	}

	// Parse the output (skip header)
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return nil, errors.New("insufficient fields in ps output")
	}

	// Check for hyphen (unavailable data)
	ioCountStr := fields[1]
	if ioCountStr == "-" {
		return nil, errors.New("I/O counters not available for this process")
	}

	// Parse the I/O count
	ioCount, err := strconv.ParseUint(ioCountStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &IOCountersStat{
		ReadBytes:  ioCount,
		WriteBytes: 0, // AIX doesn't separate read/write I/O
	}, nil
}

func (p *Process) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	// AIX does not expose context switch information via proc files or ps command.
	// According to IBM AIX documentation, the ps command field specifiers do not include
	// nvcsw (non-voluntary context switches) or vcsw (voluntary context switches).
	// These metrics are not available in the AIX proc binary structures either.
	// This metric is not available on AIX.
	return nil, common.ErrNotImplementedError
}

func (p *Process) NumFDsWithContext(ctx context.Context) (int32, error) {
	_, fnames, err := p.fillFromfdListWithContext(ctx)
	return int32(len(fnames)), err
}

func (p *Process) NumThreadsWithContext(ctx context.Context) (int32, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return p.numThreads, nil
}

func (p *Process) ThreadsWithContext(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	ret := make(map[int32]*cpu.TimesStat)
	lwpPath := common.HostProcWithContext(ctx, strconv.Itoa(int(p.Pid)), "lwp")

	tids, err := readPidsFromDir(lwpPath)
	if err != nil {
		return nil, err
	}

	for _, tid := range tids {
		_, _, cpuTimes, _, _, _, _, err := p.fillFromTIDStatWithContext(ctx, tid)
		if err != nil {
			return nil, err
		}
		ret[tid] = cpuTimes
	}

	return ret, nil
}

func (p *Process) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	_, _, cpuTimes, _, _, _, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return cpuTimes, nil
}

func (*Process) CPUAffinityWithContext(_ context.Context) ([]int32, error) {
	// AIX ps command does not support psr field specifier in System V style
	// Berkeley style ps doesn't provide CPU affinity information
	// This metric is not available on AIX
	return nil, common.ErrNotImplementedError
}

// Note: CPUPercentWithContext is NOT overridden here
// The generic implementation from process.go is used on AIX as well
// AIX ps -o %cpu can be used if needed in the future

func (p *Process) SignalsPendingWithContext(ctx context.Context) (SignalInfoStat, error) {
	// Extract pending signals from the AIXStat structure's SigPend field
	// This field is already being read from /proc/<pid>/psinfo in fillFromStatusWithContext
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return SignalInfoStat{}, err
	}

	// Read the psinfo file directly to get the SigPend field
	psInfoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(p.Pid)), "psinfo")
	psInfoFile, err := os.Open(psInfoPath)
	if err != nil {
		return SignalInfoStat{}, err
	}
	defer psInfoFile.Close()

	// Only read up to the SigPend field to avoid EOF on truncated reads
	// AIXStat starts with: Flag(4) Flag2(4) Flags(4) Nlwp(4) Stat(1) Dmodel(1) Pad1(6) = 24 bytes
	// Then SigPend which is PrSigset [4]uint64 = 32 bytes
	// Total offset to SigPend: 24 bytes
	
	// Skip the first part of the structure to get to SigPend
	var (
		flag, flag2, flags, nlwp uint32
		stat, dmodel             byte
		pad1                     [6]byte
		sigPend                  PrSigset
	)

	err = binary.Read(psInfoFile, binary.BigEndian, &flag)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &flag2)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &flags)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &nlwp)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &stat)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &dmodel)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &pad1)
	if err != nil {
		return SignalInfoStat{}, err
	}
	err = binary.Read(psInfoFile, binary.BigEndian, &sigPend)
	if err != nil {
		return SignalInfoStat{}, err
	}

	// Convert the PrSigset (which is [4]uint64) to a single uint64 for pending signals
	// The signal set uses the first 64 bits for signals 1-64 (most common)
	pendingSignals := sigPend.SsSet[0]

	return SignalInfoStat{
		PendingProcess: pendingSignals,
	}, nil
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {
	meminfo, _, err := p.fillFromStatmWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return meminfo, nil
}

func (p *Process) MemoryInfoExWithContext(ctx context.Context) (*MemoryInfoExStat, error) {
	_, memInfoEx, err := p.fillFromStatmWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return memInfoEx, nil
}

func (p *Process) PageFaultsWithContext(ctx context.Context) (*PageFaultsStat, error) {
	_, _, _, _, _, _, pageFaults, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return pageFaults, nil
}

func (*Process) ChildrenWithContext(_ context.Context) ([]*Process, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) OpenFilesWithContext(ctx context.Context) ([]OpenFilesStat, error) {
	_, ofs, err := p.fillFromfdWithContext(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]OpenFilesStat, len(ofs))
	for i, o := range ofs {
		ret[i] = *o
	}

	return ret, nil
}

func (p *Process) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	return net.ConnectionsPidWithContext(ctx, "all", p.Pid)
}

func (p *Process) ConnectionsMaxWithContext(ctx context.Context, maxConn int) ([]net.ConnectionStat, error) {
	return net.ConnectionsPidMaxWithContext(ctx, "all", p.Pid, maxConn)
}

// getConnectionsUsingNetstat retrieves network connections using AIX netstat command.
// This function is kept for backward compatibility but delegates to the net module.
//
// Deprecated: Use net module's ConnectionsPidMaxWithContext instead
func (p *Process) getConnectionsUsingNetstat(ctx context.Context, maxConn int) ([]net.ConnectionStat, error) {
	return net.ConnectionsPidMaxWithContext(ctx, "all", p.Pid, maxConn)
}

func (p *Process) MemoryMapsWithContext(ctx context.Context, _ bool) (*[]MemoryMapsStat, error) {
	// Use AIX procmap command to retrieve detailed memory address space maps
	// procmap provides information about memory regions including HEAP, STACK, TEXT, etc.
	pid := p.Pid

	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd := exec.CommandContext(ctx, "procmap", "-X", strconv.Itoa(int(pid)))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return p.parseMemoryMaps(string(output)), nil
}

// parseMemoryMaps parses procmap output and returns a list of MemoryMapsStat
// procmap -X output format:
// 1 : /etc/init
//
// Start-ADD         End-ADD               SIZE MODE  PSIZ  TYPE       VSID             MAPPED OBJECT
// 0                 10000000           262144K r--   m     KERTXT     10002
// 10000000          1000ce95               51K r-x   s     MAINTEXT   8b8117           init
// 200003d8          20036288              215K rw-   sm    MAINDATA   890192           init
func (*Process) parseMemoryMaps(output string) *[]MemoryMapsStat {
	maps := make([]MemoryMapsStat, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Start-ADD") || strings.Contains(line, ":") && !strings.HasPrefix(line, "Start") {
			// Skip header lines, empty lines, and the first line with PID info
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		mapStat := MemoryMapsStat{}

		// Parse start address (hex)
		_, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			continue
		}

		// Parse end address (hex)
		_, err = strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			continue
		}

		// Parse SIZE (e.g., "51K", "262144K", "215K")
		size := parseSizeField(fields[2])

		// MODE is in fields[3] (e.g., "r--", "r-x", "rw-")
		// PSIZ is in fields[4] (e.g., "m", "s", "sm")
		// TYPE is in fields[5] (e.g., "KERTXT", "MAINTEXT", "MAINDATA", "HEAP", "STACK", "SLIBTEXT")

		mapType := fields[5]

		// Path/name: remaining fields after VSID (fields[6])
		// The VSID is at fields[6], and mapped object starts at fields[7] or later
		pathStart := 7
		var mapPath string
		if pathStart < len(fields) {
			pathParts := fields[pathStart:]
			mapPath = strings.Join(pathParts, " ")
		}

		// Set MemoryMapsStat fields
		mapStat.Size = size
		mapStat.Rss = size // RSS approximation: use the size field
		mapStat.Path = mapPath

		// Populate descriptive path with type information if not set
		if mapStat.Path == "" {
			mapStat.Path = "[" + strings.ToLower(mapType) + "]"
		}

		maps = append(maps, mapStat)
	}

	return &maps
}

// parseSizeField converts procmap size field (e.g., "7K", "10M", "100") to bytes
func parseSizeField(sizeStr string) uint64 {
	sizeStr = strings.TrimSpace(sizeStr)

	// Check for unit suffixes
	switch {
	case strings.HasSuffix(sizeStr, "K") || strings.HasSuffix(sizeStr, "k"):
		numStr := sizeStr[:len(sizeStr)-1]
		if num, err := strconv.ParseUint(numStr, 10, 64); err == nil {
			return num * 1024
		}
	case strings.HasSuffix(sizeStr, "M") || strings.HasSuffix(sizeStr, "m"):
		numStr := sizeStr[:len(sizeStr)-1]
		if num, err := strconv.ParseUint(numStr, 10, 64); err == nil {
			return num * 1024 * 1024
		}
	case strings.HasSuffix(sizeStr, "G") || strings.HasSuffix(sizeStr, "g"):
		numStr := sizeStr[:len(sizeStr)-1]
		if num, err := strconv.ParseUint(numStr, 10, 64); err == nil {
			return num * 1024 * 1024 * 1024
		}
	}

	// No suffix, try to parse as plain number (bytes)
	if num, err := strconv.ParseUint(sizeStr, 10, 64); err == nil {
		return num
	}

	return 0
}

func (*Process) EnvironWithContext(_ context.Context) ([]string, error) {
	// AIX /proc does not expose environment variables in a standard text format
	// Envp in psinfo is a user-space pointer that is not directly accessible
	return nil, common.ErrNotImplementedError
}

/**
** Internal functions
**/

func limitToUint(val string) (uint64, error) {
	if val == "unlimited" {
		return math.MaxUint64, nil
	}
	res, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return res, nil
}

// Get num_fds from /proc/(pid)/limits (not available in AIX)
func (*Process) fillFromLimitsWithContext(_ context.Context) ([]RlimitStat, error) {
	// AIX /proc does not expose resource limits in a standard procfs location
	return nil, common.ErrNotImplementedError
}

// Get list of /proc/(pid)/fd files
func (p *Process) fillFromfdListWithContext(ctx context.Context) (string, []string, error) {
	pid := p.Pid
	statPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "fd")
	d, err := os.Open(statPath)
	if err != nil {
		return statPath, []string{}, err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	return statPath, fnames, err
}

// Get num_fds from /proc/(pid)/fd
func (p *Process) fillFromfdWithContext(ctx context.Context) (int32, []*OpenFilesStat, error) {
	statPath, fnames, err := p.fillFromfdListWithContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	numFDs := int32(len(fnames))

	var openfiles []*OpenFilesStat
	for _, fd := range fnames {
		fpath := filepath.Join(statPath, fd)
		linkPath, err := os.Readlink(fpath)
		if err != nil {
			continue
		}
		t, err := strconv.ParseUint(fd, 10, 64)
		if err != nil {
			return numFDs, openfiles, err
		}
		o := &OpenFilesStat{
			Path: linkPath,
			Fd:   t,
		}
		openfiles = append(openfiles, o)
	}

	return numFDs, openfiles, nil
}

// Get cwd from /proc/(pid)/cwd
func (p *Process) fillFromCwdWithContext(ctx context.Context) (string, error) {
	pid := p.Pid
	cwdPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "cwd")
	cwd, err := os.Readlink(cwdPath)
	if err != nil {
		return "", err
	}
	return string(cwd), nil
}

// Get exe from /proc/(pid)/psinfo
func (p *Process) fillFromExeWithContext(ctx context.Context) (string, error) {
	pid := p.Pid
	infoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	infoFile, err := os.Open(infoPath)
	if err != nil {
		return "", err
	}
	defer infoFile.Close()

	var aixPSinfo AIXPSInfo
	err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
	if err != nil {
		return "", err
	}

	// Try Fname field from psinfo first (most reliable)
	fname := extractFnameString(&aixPSinfo)
	if fname != "" {
		return fname, nil
	}

	// Fallback to extracting from Psargs field
	psargs := extractPsargsString(&aixPSinfo)
	if psargs != "" {
		// Extract the first word (executable name) from Psargs
		parts := strings.Fields(psargs)
		if len(parts) > 0 {
			return filepath.Base(parts[0]), nil
		}
	}

	// Get first argument from process address space (argv[0] is the executable name)
	// This is last resort as it's more prone to permission/corruption issues
	args, err := p.readArgsFromAddressSpace(ctx, int(pid), &aixPSinfo, 1)
	if err == nil && len(args) > 0 {
		// Extract just the basename from the full path
		exeName := filepath.Base(args[0])
		return exeName, nil
	}

	return "", nil
}

// getProcessBitness returns the pointer size (4 or 8 bytes) for a process, caching the result
func (p *Process) getProcessBitness(ctx context.Context, pid int) (int64, error) {
	// Return cached value if available
	if cached, ok := aixBitnessCache.Load(p.Pid); ok {
		return cached.(int64), nil
	}

	// Read status file to get data model byte (offset 17 in AIXStat structure)
	statusPath := common.HostProcWithContext(ctx, strconv.Itoa(pid), "status")
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		return 8, err // Default to 64-bit on error
	}

	ptrSize := int64(8) // Default to 64-bit
	if len(statusData) > 17 {
		if statusData[17] == 0 {
			ptrSize = 4 // 32-bit process (byte 0 indicates 32-bit)
		}
	}

	// Cache the bitness
	aixBitnessCache.Store(p.Pid, ptrSize)

	return ptrSize, nil
}

// readArgsFromAddressSpace reads argument and environment strings from process memory
// Similar to OSHI's approach
func (p *Process) readArgsFromAddressSpace(ctx context.Context, pid int, psinfo *AIXPSInfo, maxArgs int) ([]string, error) {
	if psinfo.Argc == 0 || psinfo.Argc > 10000 {
		// Sanity check on argc
		return nil, common.ErrNotImplementedError
	}

	asPath := common.HostProcWithContext(ctx, strconv.Itoa(pid), "as")
	fd, err := syscall.Open(asPath, syscall.O_RDONLY, 0)
	if err != nil {
		// No permission or file not found
		return nil, err
	}
	defer syscall.Close(fd)

	// Get cached bitness (pointer size)
	ptrSize, err := p.getProcessBitness(ctx, pid)
	if err != nil {
		// If we can't determine bitness, default to 64-bit
		ptrSize = 8
	}

	// Read argv pointers
	argc := int(psinfo.Argc)
	if argc > maxArgs && maxArgs > 0 {
		argc = maxArgs
	}

	argv := make([]int64, argc)
	for i := 0; i < argc; i++ {
		offset := int64(psinfo.Argv) + int64(i)*ptrSize
		buf := make([]byte, ptrSize)
		n, err := syscall.Pread(fd, buf, offset)
		if err != nil || n != len(buf) {
			break
		}
		if ptrSize == 8 {
			argv[i] = int64(binary.BigEndian.Uint64(buf))
		} else {
			argv[i] = int64(binary.BigEndian.Uint32(buf))
		}
	}

	// Read argument strings
	args := make([]string, 0, argc)
	for i := 0; i < argc && i < len(argv); i++ {
		if argv[i] == 0 {
			break
		}
		argStr, err := readStringFromAddressSpace(fd, argv[i])
		if err != nil {
			break
		}
		if argStr != "" {
			args = append(args, argStr)
		}
	}

	return args, nil
}

// readStringFromAddressSpace reads a null-terminated string from process memory
func readStringFromAddressSpace(fd int, addr int64) (string, error) {
	const pageSize = 4096
	const maxStrLen = 32768

	// Align to page boundary
	pageStart := (addr / pageSize) * pageSize
	buffer := make([]byte, pageSize*2)

	n, err := syscall.Pread(fd, buffer, pageStart)
	if err != nil || n == 0 {
		return "", err
	}

	// Calculate offset within buffer
	offset := addr - pageStart
	if offset < 0 || offset >= int64(len(buffer)) {
		return "", common.ErrNotImplementedError
	}

	// Read null-terminated string
	var result strings.Builder
	for i := offset; i < int64(len(buffer)) && i < offset+int64(maxStrLen); i++ {
		if buffer[i] == 0 {
			break
		}
		result.WriteByte(buffer[i])
	}

	return result.String(), nil
}

// Get cmdline from /proc/(pid)/psinfo by reading from address space
func (p *Process) fillFromCmdlineWithContext(ctx context.Context) (string, error) {
	pid := p.Pid
	infoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	infoFile, err := os.Open(infoPath)
	if err != nil {
		return "", err
	}
	defer infoFile.Close()

	var aixPSinfo AIXPSInfo
	err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
	if err != nil {
		return "", err
	}

	// Use Psargs field directly - it contains the initial command line
	psargs := extractPsargsString(&aixPSinfo)
	if psargs != "" {
		return psargs, nil
	}

	// If Psargs is empty, try reading from address space
	args, err := p.readArgsFromAddressSpace(ctx, int(pid), &aixPSinfo, 0)
	if err == nil && len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	return "", nil
}

func (p *Process) fillSliceFromCmdlineWithContext(ctx context.Context) ([]string, error) {
	pid := p.Pid
	infoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	infoFile, err := os.Open(infoPath)
	if err != nil {
		return nil, err
	}
	defer infoFile.Close()

	var aixPSinfo AIXPSInfo
	err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
	if err != nil {
		return nil, err
	}

	// Use Psargs field directly - it contains the initial command line
	psargs := extractPsargsString(&aixPSinfo)
	if psargs != "" {
		// Split on spaces as a simple heuristic; AIX psinfo.Psargs is limited to 80 chars
		return strings.Fields(psargs), nil
	}

	// If Psargs is empty, try reading arguments directly from address space
	args, err := p.readArgsFromAddressSpace(ctx, int(pid), &aixPSinfo, 0)
	if err == nil && len(args) > 0 {
		return args, nil
	}

	return []string{}, nil
}

// extractPsargsString extracts and cleans the Psargs field from AIXPSInfo
func extractPsargsString(psinfo *AIXPSInfo) string {
	return string(bytes.TrimRight(psinfo.Psargs[:], "\x00"))
}

// extractFnameString extracts and cleans the Fname field from AIXPSInfo
// Fname may have leading null bytes, so we need to skip them first
func extractFnameString(psinfo *AIXPSInfo) string {
	// First, trim trailing null bytes
	trimmed := bytes.TrimRight(psinfo.Fname[:], "\x00")
	// Then, trim leading null bytes
	trimmed = bytes.TrimLeft(trimmed, "\x00")
	return string(trimmed)
}

// Get IO status from /proc/(pid)/status (not available in AIX)
func (*Process) fillFromIOWithContext(_ context.Context) (*IOCountersStat, error) {
	// AIX does not expose detailed I/O counters in /proc; return nil
	return nil, common.ErrNotImplementedError
}

// Get memory info from /proc/(pid)/psinfo
func (p *Process) fillFromStatmWithContext(ctx context.Context) (*MemoryInfoStat, *MemoryInfoExStat, error) {
	pid := p.Pid
	infoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	infoFile, err := os.Open(infoPath)
	if err != nil {
		return nil, nil, err
	}
	defer infoFile.Close()

	var aixPSinfo AIXPSInfo
	err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
	if err != nil {
		return nil, nil, err
	}

	// Read memory from AIXPSInfo.Size and AIXPSInfo.Rssize fields (matching OSHI)
	// These are in KB, multiply by 1024 to get bytes
	vms := aixPSinfo.Size * 1024
	rss := aixPSinfo.Rssize * 1024

	meminfo := &MemoryInfoStat{
		VMS: vms,
		RSS: rss,
	}
	meminfoEx := &MemoryInfoExStat{
		VMS: vms,
		RSS: rss,
	}
	return meminfo, meminfoEx, nil
}

// Get name from /proc/(pid)/psinfo (Fname field)
func (p *Process) fillFromCommWithContext(ctx context.Context) error {
	exe, err := p.fillFromExeWithContext(ctx)
	if err != nil {
		return err
	}
	p.name = exe
	return nil
}

// Get various status from /proc/(pid)/status
func (p *Process) fillFromStatus() error {
	return p.fillFromStatusWithContext(context.Background())
}

func (p *Process) fillFromStatusWithContext(ctx context.Context) error {
	pid := p.Pid
	statusPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "status")
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return err
	}
	defer statusFile.Close()

	// Parse the binary AIXStat structure
	var aixStat AIXStat
	err = binary.Read(statusFile, binary.BigEndian, &aixStat)
	if err != nil {
		return err
	}

	p.numCtxSwitches = &NumCtxSwitchesStat{}
	p.memInfo = &MemoryInfoStat{}
	p.sigInfo = &SignalInfoStat{}

	// Extract process state
	p.status = convertStatusChar(string([]byte{aixStat.Stat}))
	// Recognize AIX-specific status codes if the converted value is empty
	if p.status == "" {
		// Status byte not recognized - use AIX-specific status codes if needed
		switch aixStat.Stat {
		case 0:
			p.status = "NONE"
		case 1:
			p.status = Running // SACTIVE
		case 2:
			p.status = Sleep // SSLEEP
		case 3:
			p.status = Stop // SSTOP
		case 4:
			p.status = Zombie // SZOMB
		case 5:
			p.status = Idle // SIDL
		case 6:
			p.status = Wait // SWAIT
		case 7:
			p.status = Running // SORPHAN - treat as running
		default:
			p.status = UnknownState
		}
	}

	// Extract parent PID
	p.parent = int32(aixStat.Ppid)

	// Extract TGID (same as PID on AIX, as there's no separate TGID concept)
	p.tgid = int32(aixStat.Pid)

	// Cache bitness: dmodel field indicates 32-bit (0) or 64-bit (non-zero)
	if aixStat.Dmodel == 0 {
		aixBitnessCache.Store(p.Pid, int64(4))
	} else {
		aixBitnessCache.Store(p.Pid, int64(8))
	}

	// Also read psinfo for UID/GID and thread count
	infoPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	infoFile, err := os.Open(infoPath)
	if err == nil {
		defer infoFile.Close()
		var aixPSinfo AIXPSInfo
		err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
		if err == nil {
			// Extract UIDs: real UID, effective UID, saved UID (use effective as third), and fsuid (use effective)
			p.uids = []uint32{uint32(aixPSinfo.UID), uint32(aixPSinfo.Euid), uint32(aixPSinfo.Euid), uint32(aixPSinfo.Euid)}
			// Extract GIDs: real GID, effective GID, saved GID (use effective as third), and fsgid (use effective)
			p.gids = []uint32{uint32(aixPSinfo.Gid), uint32(aixPSinfo.Egid), uint32(aixPSinfo.Egid), uint32(aixPSinfo.Egid)}
			// Extract number of threads from Nlwp field
			p.numThreads = int32(aixPSinfo.Nlwp)
		}
	}

	return nil
}

func (p *Process) fillFromTIDStat(tid int32) (uint64, int32, *cpu.TimesStat, int64, uint32, int32, *PageFaultsStat, error) {
	return p.fillFromTIDStatWithContext(context.Background(), tid)
}

func (p *Process) fillFromTIDStatWithContext(ctx context.Context, tid int32) (uint64, int32, *cpu.TimesStat, int64, uint32, int32, *PageFaultsStat, error) {
	pid := p.Pid
	var statPath string
	var infoPath string
	var lwpStatPath string
	var lwpInfoPath string
	var lwpStatFile *os.File
	var lwpInfoFile *os.File

	statPath = common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "status")
	infoPath = common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo")
	if tid > -1 {
		// Search for lwpstatus and lwpinfo files, handling unknown directory structures
		tidStr := strconv.Itoa(int(tid))
		basePath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "lwp", tidStr)
		lwpStatPath, lwpInfoPath = findLwpFiles(basePath)
	}

	// Open the binary files
	statFile, err := os.Open(statPath)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	defer statFile.Close()
	infoFile, err := os.Open(infoPath)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	defer infoFile.Close()
	if tid > -1 {
		var err error
		lwpStatFile, err = os.Open(lwpStatPath)
		if err != nil {
			// If we can't open lwp files, just use the main process files (tid = -1 behavior)
			// This is a graceful fallback for processes without thread info
			tid = -1
		} else {
			defer lwpStatFile.Close()
			lwpInfoFile, err = os.Open(lwpInfoPath)
			if err != nil {
				// If we can't open lwp info, close the stat file and fall back
				lwpStatFile.Close()
				tid = -1
			} else {
				defer lwpInfoFile.Close()
			}
		}
	}

	// We need to read a few binary files into a struct variables
	var aixStat AIXStat
	var aixPSinfo AIXPSInfo
	var aixlwpStat LwpStatus
	var aixlspPSinfo LwpsInfo
	err = binary.Read(statFile, binary.BigEndian, &aixStat)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	err = binary.Read(infoFile, binary.BigEndian, &aixPSinfo)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	if tid > -1 {
		err = binary.Read(lwpStatFile, binary.BigEndian, &aixlwpStat)
		if err != nil {
			return 0, 0, nil, 0, 0, 0, nil, err
		}
		err = binary.Read(lwpInfoFile, binary.BigEndian, &aixlspPSinfo)
		if err != nil {
			return 0, 0, nil, 0, 0, 0, nil, err
		}
	}

	// TODO: Figure out how to get terminal information for this process

	ppid := aixStat.Ppid
	utime := float64(aixStat.Utime.TvSec)
	stime := float64(aixStat.Stime.TvSec)

	iotime := 0.0 // TODO: Figure out actual iotime for AIX

	cpuTimes := &cpu.TimesStat{
		CPU:    "cpu",
		User:   utime / float64(clockTicks),
		System: stime / float64(clockTicks),
		Iowait: iotime / float64(clockTicks),
	}

	bootTime, _ := common.BootTimeWithContext(ctx, invoke)
	startTime := uint64(aixPSinfo.Start.TvSec)
	createTime := int64((startTime * 1000 / uint64(clockTicks)) + uint64(bootTime*1000))

	// This information is only available at thread level
	var rtpriority uint32
	var nice int32
	if tid > -1 {
		rtpriority = uint32(aixlspPSinfo.Pri)
		nice = int32(aixlspPSinfo.Nice)
	}

	// Extract page fault data via ps command for more detailed info
	pageFaults, _ := p.getPageFaults(ctx)

	return 0, int32(ppid), cpuTimes, createTime, uint32(rtpriority), nice, pageFaults, nil
}

// getPageFaults retrieves page fault information for the process
func (p *Process) getPageFaults(ctx context.Context) (*PageFaultsStat, error) {
	// Query page faults via ps command
	//nolint:gosec // Process ID from internal tracking, not untrusted input
	cmd := exec.CommandContext(ctx, "ps", "-v", "-p", strconv.Itoa(int(p.Pid)))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, errors.New("insufficient ps output")
	}

	// Parse ps v output - look for page fault related columns
	// AIX ps -v output includes: PID, USER, TTY, STAT, TIME, SZ, RSS, %MEM, PAGEIN, etc.
	fields := strings.Fields(lines[1])

	// Try to extract PAGEIN (major page faults)
	pageFaults := &PageFaultsStat{}

	// Look for numeric fields that indicate page faults
	// The exact column varies, so we'll try a heuristic approach
	if len(fields) >= 9 {
		if pagein, err := strconv.ParseUint(fields[len(fields)-1], 10, 64); err == nil {
			pageFaults.MajorFaults = pagein
		}
	}

	return pageFaults, nil
}

func (p *Process) fillFromStatWithContext(ctx context.Context) (uint64, int32, *cpu.TimesStat, int64, uint32, int32, *PageFaultsStat, error) {
	return p.fillFromTIDStatWithContext(ctx, -1)
}

func pidsWithContext(ctx context.Context) ([]int32, error) {
	return readPidsFromDir(common.HostProcWithContext(ctx))
}

func ProcessesWithContext(ctx context.Context) ([]*Process, error) {
	out := []*Process{}

	pids, err := PidsWithContext(ctx)
	if err != nil {
		return out, err
	}

	for _, pid := range pids {
		p, err := NewProcessWithContext(ctx, pid)
		if err != nil {
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

func readPidsFromDir(path string) ([]int32, error) {
	var ret []int32

	d, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fnames, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, fname := range fnames {
		pid, err := strconv.ParseInt(fname, 10, 32)
		if err != nil {
			// if not numeric name, just skip
			continue
		}
		ret = append(ret, int32(pid))
	}

	return ret, nil
}

func splitProcStat(content []byte) []string {
	nameStart := bytes.IndexByte(content, '(')
	nameEnd := bytes.LastIndexByte(content, ')')

	// Defensive checks for malformed input
	if nameStart < 0 || nameEnd < 0 || nameStart >= nameEnd {
		// Malformed input; return empty result to avoid panic
		return []string{}
	}

	// Ensure rest offset is within bounds
	restStart := nameEnd + 2
	if restStart > len(content) {
		restStart = len(content)
	}

	restFields := strings.Fields(string(content[restStart:])) // +2 skip ') '
	name := content[nameStart+1 : nameEnd]
	pid := strings.TrimSpace(string(content[:nameStart]))
	fields := make([]string, 3, len(restFields)+3)
	fields[1] = string(pid)
	fields[2] = string(name)
	fields = append(fields, restFields...)
	return fields
}

// extractString extracts a null-terminated string from a byte slice,
// handling non-printable characters gracefully
func extractString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			// Found null terminator, return up to here
			return string(b[:i])
		}
	}
	// No null terminator, return all bytes after trimming null bytes from the end
	return strings.TrimRight(string(b), "\x00")
}

// findLwpFiles searches recursively for lwpstatus and lwpinfo files under a given directory.
// AIX /proc/pid/lwp structure can vary, so this function explores the directory tree to find the files.
func findLwpFiles(basePath string) (string, string) {
	// First try the direct path: basePath/lwpstatus
	directStatPath := filepath.Join(basePath, "lwpstatus")
	directInfoPath := filepath.Join(basePath, "lwpinfo")
	if _, err := os.Stat(directStatPath); err == nil {
		return directStatPath, directInfoPath
	}

	// If direct path doesn't exist, recursively search the directory tree
	statPath, infoPath := searchLwpFilesRecursive(basePath)
	return statPath, infoPath
}

// searchLwpFilesRecursive recursively searches for lwpstatus and lwpinfo files.
func searchLwpFilesRecursive(searchPath string) (string, string) {
	d, err := os.Open(searchPath)
	if err != nil {
		return "", ""
	}
	defer d.Close()

	entries, err := d.Readdirnames(-1)
	if err != nil {
		return "", ""
	}

	var statPath, infoPath string

	for _, entry := range entries {
		fullPath := filepath.Join(searchPath, entry)

		// Check if this entry is one of our target files
		if entry == "lwpstatus" && statPath == "" {
			statPath = fullPath
		}
		if entry == "lwpinfo" && infoPath == "" {
			infoPath = fullPath
		}

		// If we found both files, return immediately
		if statPath != "" && infoPath != "" {
			return statPath, infoPath
		}

		// Check if this is a directory and recurse into it
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			foundStatPath, foundInfoPath := searchLwpFilesRecursive(fullPath)
			if foundStatPath != "" && statPath == "" {
				statPath = foundStatPath
			}
			if foundInfoPath != "" && infoPath == "" {
				infoPath = foundInfoPath
			}
			// If we found both files, return immediately
			if statPath != "" && infoPath != "" {
				return statPath, infoPath
			}
		}
	}

	return statPath, infoPath
}
