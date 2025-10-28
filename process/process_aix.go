// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/internal/common"
	"github.com/shirou/gopsutil/v4/net"
)

var pageSize = uint64(os.Getpagesize())

const prioProcess = 0 // linux/resource.h

var clockTicks = 100 // default value

func init() {
	// TODO: Once the sysconf library supports AIX, use the commented code below
	// clkTck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	// ignore errors
	// if err == nil {
	// 	clockTicks = int(clkTck)
	// }

	// For now, this will use the hard coded default of 100
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
	SiUid    uint64    // real user id of sending process
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
	Wtype   uint16    // type of thread wait
	State   byte      // thread state
	Sname   byte      // printable thread state character
	Nice    uint16    // nice value for CPU usage
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
	Uid    uint32         // real user id
	Euid   uint32         // effective user id
	Gid    uint32         // real group id
	Egid   uint32         // effective group id
	Argc   uint32         // initial argument count
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
		if err := p.fillNameWithContext(ctx); err != nil {
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

func (p *Process) ForegroundWithContext(ctx context.Context) (bool, error) {
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
	return "", common.ErrNotImplementedError
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	_, _, _, _, _, nice, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return nice, nil
}

func (p *Process) IOniceWithContext(ctx context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	return p.RlimitUsageWithContext(ctx, false)
}

func (p *Process) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	return p.fillFromIOWithContext(ctx)
}

func (p *Process) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	err := p.fillFromStatusWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return p.numCtxSwitches, nil
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

func (p *Process) CPUAffinityWithContext(ctx context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
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

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
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

func (p *Process) MemoryMapsWithContext(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) EnvironWithContext(ctx context.Context) ([]string, error) {
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

// Get num_fds from /proc/(pid)/limits
func (p *Process) fillFromLimitsWithContext(ctx context.Context) ([]RlimitStat, error) {
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
		filepath, err := os.Readlink(fpath)
		if err != nil {
			continue
		}
		t, err := strconv.ParseUint(fd, 10, 64)
		if err != nil {
			return numFDs, openfiles, err
		}
		o := &OpenFilesStat{
			Path: filepath,
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

// Get exe from /proc/(pid)/exe
func (p *Process) fillFromExeWithContext(ctx context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

// Get cmdline from /proc/(pid)/cmdline
func (p *Process) fillFromCmdlineWithContext(ctx context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func (p *Process) fillSliceFromCmdlineWithContext(ctx context.Context) ([]string, error) {
	return nil, common.ErrNotImplementedError
}

// Get IO status from /proc/(pid)/io
func (p *Process) fillFromIOWithContext(ctx context.Context) (*IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

// Get memory info from /proc/(pid)/statm
func (p *Process) fillFromStatmWithContext(ctx context.Context) (*MemoryInfoStat, *MemoryInfoExStat, error) {
	return nil, nil, common.ErrNotImplementedError
}

// Get name from /proc/(pid)/comm or /proc/(pid)/status
func (p *Process) fillNameWithContext(ctx context.Context) error {
	err := p.fillFromCommWithContext(ctx)
	if err == nil && p.name != "" && len(p.name) < 15 {
		return nil
	}
	return p.fillFromStatusWithContext(ctx)
}

// Get name from /proc/(pid)/comm
func (p *Process) fillFromCommWithContext(ctx context.Context) error {
	return common.ErrNotImplementedError
}

// Get various status from /proc/(pid)/status
func (p *Process) fillFromStatus() error {
	return p.fillFromStatusWithContext(context.Background())
}

func (p *Process) fillFromStatusWithContext(ctx context.Context) error {
	pid := p.Pid
	statPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "status")
	contents, err := os.ReadFile(statPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contents), "\n")
	p.numCtxSwitches = &NumCtxSwitchesStat{}
	p.memInfo = &MemoryInfoStat{}
	p.sigInfo = &SignalInfoStat{}
	for _, line := range lines {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		value := tabParts[1]
		switch strings.TrimRight(tabParts[0], ":") {
		case "Name":
			p.name = strings.Trim(value, " \t")
			if len(p.name) >= 15 {
				cmdlineSlice, err := p.CmdlineSliceWithContext(ctx)
				if err != nil {
					return err
				}
				if len(cmdlineSlice) > 0 {
					extendedName := filepath.Base(cmdlineSlice[0])
					if strings.HasPrefix(extendedName, p.name) {
						p.name = extendedName
					}
				}
			}
			// Ensure we have a copy and not reference into slice
			p.name = string([]byte(p.name))
		case "State":
			p.status = convertStatusChar(value[0:1])
			// Ensure we have a copy and not reference into slice
			p.status = string([]byte(p.status))
		case "PPid", "Ppid":
			pval, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			p.parent = int32(pval)
		case "Tgid":
			pval, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			p.tgid = int32(pval)
		case "Uid":
			p.uids = make([]uint32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					return err
				}
				p.uids = append(p.uids, uint32(v))
			}
		case "Gid":
			p.gids = make([]uint32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					return err
				}
				p.gids = append(p.gids, uint32(v))
			}
		case "Groups":
			groups := strings.Fields(value)
			p.groups = make([]uint32, 0, len(groups))
			for _, i := range groups {
				v, err := strconv.ParseUint(i, 10, 32)
				if err != nil {
					return err
				}
				p.groups = append(p.groups, uint32(v))
			}
		case "Threads":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			p.numThreads = int32(v)
		case "voluntary_ctxt_switches":
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			p.numCtxSwitches.Voluntary = v
		case "nonvoluntary_ctxt_switches":
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			p.numCtxSwitches.Involuntary = v
		case "VmRSS":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.RSS = v * 1024
		case "VmSize":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.VMS = v * 1024
		case "VmSwap":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.Swap = v * 1024
		case "VmHWM":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.HWM = v * 1024
		case "VmData":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.Data = v * 1024
		case "VmStk":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.Stack = v * 1024
		case "VmLck":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			p.memInfo.Locked = v * 1024
		case "SigPnd":
			if len(value) > 16 {
				value = value[len(value)-16:]
			}
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				return err
			}
			p.sigInfo.PendingThread = v
		case "ShdPnd":
			if len(value) > 16 {
				value = value[len(value)-16:]
			}
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				return err
			}
			p.sigInfo.PendingProcess = v
		case "SigBlk":
			if len(value) > 16 {
				value = value[len(value)-16:]
			}
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				return err
			}
			p.sigInfo.Blocked = v
		case "SigIgn":
			if len(value) > 16 {
				value = value[len(value)-16:]
			}
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				return err
			}
			p.sigInfo.Ignored = v
		case "SigCgt":
			if len(value) > 16 {
				value = value[len(value)-16:]
			}
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				return err
			}
			p.sigInfo.Caught = v
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
		lwpStatPath = common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "lwp", strconv.Itoa(int(tid)), strconv.Itoa(int(tid)), "lwpstatus")
		lwpInfoPath = common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "lwp", strconv.Itoa(int(tid)), strconv.Itoa(int(tid)), "lwpinfo")
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
			return 0, 0, nil, 0, 0, 0, nil, err
		}
		defer lwpStatFile.Close()
		lwpInfoFile, err = os.Open(lwpInfoPath)
		if err != nil {
			return 0, 0, nil, 0, 0, 0, nil, err
		}
		defer lwpInfoFile.Close()
	}

	// We need to read a few binary files into a struct variables
	var aixStat AIXStat
	var aixPSinfo AIXPSInfo
	var aixlwpStat LwpStatus
	var aixlspPSinfo LwpsInfo
	err = binary.Read(statFile, binary.LittleEndian, &aixStat)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	err = binary.Read(infoFile, binary.LittleEndian, &aixPSinfo)
	if err != nil {
		return 0, 0, nil, 0, 0, 0, nil, err
	}
	if tid > -1 {
		err = binary.Read(statFile, binary.LittleEndian, &aixlwpStat)
		if err != nil {
			return 0, 0, nil, 0, 0, 0, nil, err
		}
		err = binary.Read(infoFile, binary.LittleEndian, &aixlspPSinfo)
		if err != nil {
			return 0, 0, nil, 0, 0, 0, nil, err
		}
	}

	// TODO: Figure out how to get terminal information for this process

	ppid := aixStat.Ppid
	utime := float64(aixStat.Utime.TvSec)
	stime := float64(aixStat.Stime.TvSec)

	var iotime float64
	iotime = 0 // TODO: Figure out actual iotime for AIX

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

	return 0, int32(ppid), cpuTimes, createTime, uint32(rtpriority), nice, nil, nil
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
	restFields := strings.Fields(string(content[nameEnd+2:])) // +2 skip ') '
	name := content[nameStart+1 : nameEnd]
	pid := strings.TrimSpace(string(content[:nameStart]))
	fields := make([]string, 3, len(restFields)+3)
	fields[1] = string(pid)
	fields[2] = string(name)
	fields = append(fields, restFields...)
	return fields
}
