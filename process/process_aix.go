// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
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

type pr_timestruc64_t struct {
	tv_sec  int64  /* 64 bit time_t value      */
	tv_nsec int32  /* 32 bit suseconds_t value */
	__pad   uint32 /* reserved for future use  */
}

/* hardware fault set */
type fltset struct {
	flt_set [4]uint64 /* fault set */
}

type pr_sigset struct {
	ss_set [4]uint64 /* signal set */
}

type prptr64 uint64

type size64 uint64

type pid64 uint64 /* size invariant 64-bit pid  */

type pr_siginfo64 struct {
	si_signo  int32     /* signal number */
	si_errno  int32     /* if non-zero, errno of this signal */
	si_code   int32     /* signal code */
	si_imm    int32     /* immediate data */
	si_status int32     /* exit value or signal */
	__pad1    uint32    /* reserved for future use */
	si_uid    uint64    /* real user id of sending process */
	si_pid    uint64    /* sending process id */
	si_addr   prptr64   /* address of faulting instruction */
	si_band   int64     /* band event for SIGPOLL */
	si_value  prptr64   /* signal value */
	__pad     [4]uint32 /* reserved for future use */
}

type pr_sigaction64 struct {
	sa_union prptr64
	sa_mask  pr_sigset /* signal mask */
	sa_flags int32     /* signal flags */
	__pad    [5]int32  /* reserved for future use */
}

type pr_stack64 struct {
	ss_sp    prptr64  /* stack base or pointer */
	ss_size  uint64   /* stack size */
	ss_flags int32    /* flags */
	__pad    [5]int32 /* reserved for future use */
}

type pr_timestruc64 struct {
	tv_sec  int64  /* 64 bit time_t value      */
	tv_nsec int32  /* 32 bit suseconds_t value */
	__pad   uint32 /* reserved for future use  */
}

type prgregset struct {
	__iar    size64     /* Instruction Pointer          */
	__msr    size64     /* machine state register       */
	__cr     size64     /* condition register           */
	__lr     size64     /* link register                */
	__ctr    size64     /* count register               */
	__xer    size64     /* fixed point exception        */
	__fpscr  size64     /* floating point status reg    */
	__fpscrx size64     /* extension floating point     */
	__gpr    [32]size64 /* static general registers     */
	__USPRG3 size64
	__pad1   [7]size64 /* Reserved for future use      */
}

type prfpregset struct {
	__fpr [32]float64 /* Floating Point Registers        */

}

type pfamily struct {
	/* offset of the prextset relative to the
	* current status file (status or lwpstatus)
	* zero if none */
	pr_extoff  uint64
	pr_extsize uint64     /* size of extension, zero if none */
	pr_pad     [14]uint64 /* reserved for future use.  Sized to allow
	 * for 14 64-bit registers */
}

type lwpStatus struct {
	pr_lwpid    uint64         /* specific thread id */
	pr_flags    uint32         /* thread status flags */
	pr__pad1    [1]byte        /* reserved for future use */
	pr_state    byte           /* thread state */
	pr_cursig   uint16         /* current signal */
	pr_why      uint16         /* stop reason */
	pr_what     uint16         /* more detailed reason */
	pr_policy   uint32         /* scheduling policy */
	pr_clname   [8]byte        /* printable scheduling policy string */
	pr_lwppend  pr_sigset      /* set of signals pending to the thread */
	pr_lwphold  pr_sigset      /* set of signals blocked by the thread */
	pr_info     pr_siginfo64   /* info associated with signal or fault */
	pr_altstack pr_stack64     /* alternate signal stack info */
	pr_action   pr_sigaction64 /* signal action for current signal */
	pr__pad2    uint32         /* reserved for future use */
	pr_syscall  uint16         /* system call number (if in syscall) */
	/* number of arguments to this syscall
	 * syscall arguments,
	 * pr_sysarg[0] contains syscall return
	 * value when stopped on PR_SYSEXIT event */
	pr_nsysarg uint16
	pr_sysarg  [8]uint64
	pr_errno   int32  /* errno from last syscall */
	pr_ptid    uint32 /* pthread id */

	pr__pad   [9]uint64  /* reserved for future use */
	pr_reg    prgregset  /* general and special registers */
	pr_fpreg  prfpregset /* floating point registers */
	pr_family pfamily    /* hardware platform specific information */
}

type AIXStat struct {
	pr_flag     uint32           /* process flags from proc struct p_flag */
	pr_flag2    uint32           /* process flags from proc struct p_flag2 */
	pr_flags    uint32           /* /proc flags */
	pr_nlwp     uint32           /* number of threads in the process */
	pr_stat     byte             /* process state from proc p_stat */
	pr_dmodel   byte             /* data model for the process */
	pr__pad1    [6]byte          /* reserved for future use */
	pr_sigpend  pr_sigset        /* set of process pending signals */
	pr_brkbase  prptr64          /* address of the process heap */
	pr_brksize  uint64           /* size of the process heap, in bytes */
	pr_stkbase  prptr64          /* address of the process stack */
	pr_stksize  uint64           /* size of the process stack, in bytes */
	pr_pid      pid64            /* process id */
	pr_ppid     pid64            /* parent process id */
	pr_pgid     pid64            /* process group id */
	pr_sid      pid64            /* session id */
	pr_utime    pr_timestruc64_t /* process user cpu time */
	pr_stime    pr_timestruc64_t /* process system cpu time */
	pr_cutime   pr_timestruc64_t /* sum of children's user times */
	pr_cstime   pr_timestruc64_t /* sum of children's system times */
	pr_sigtrace pr_sigset        /* mask of traced signals */
	pr_flttrace fltset           /* mask of traced hardware faults */
	/* offset into pstatus file of sysset_t
	 * identifying system calls traced on
	 * entry.  If 0, then no entry syscalls
	 * are being traced. */
	pr_sysentry_offset uint32
	/* offset into pstatus file of sysset_t
	 * identifying system calls traced on
	 * exit.  If 0, then no exit syscalls
	 * are being traced. */
	pr_sysexit_offset uint32
	pr__pad           [8]uint64 /* reserved for future use */
	pr_lwp            lwpStatus /* "representative" thread status */
}

type lwpsInfo struct {
	pr_lwpid   uint64    /* thread id */
	pr_addr    uint64    /* internal address of thread */
	pr_wchan   uint64    /* wait address for sleeping thread */
	pr_flag    uint32    /* thread flags */
	pr_wtype   uint16    /* type of thread wait */
	pr_state   byte      /* thread state */
	pr_sname   byte      /* printable thread state character */
	pr_nice    uint16    /* nice value for CPU usage */
	pr_pri     int32     /* priority, high value = high priority */
	pr_policy  uint32    /* scheduling policy */
	pr_clname  [8]byte   /* printable scheduling policy string */
	pr_onpro   int32     /* processor on which thread last ran */
	pr_bindpro int32     /* processor to which thread is bound */
	pr_ptid    uint32    /* pthread id */
	pr__pad1   uint32    /* reserved for future use */
	pr__pad    [7]uint64 /* reserved for future use */
}

type aixPSInfo struct {
	pr_flag   uint32           /* process flags from proc struct p_flag */
	pr_flag2  uint32           /* process flags from proc struct p_flag2 */
	pr_nlwp   uint32           /* number of threads in process */
	pr_uid    uint32           /* real user id */
	pr_euid   uint32           /* effective user id */
	pr_gid    uint32           /* real group id */
	pr_egid   uint32           /* effective group id */
	pr_argc   uint32           /* initial argument count */
	pr_pid    uint64           /* unique process id */
	pr_ppid   uint64           /* process id of parent */
	pr_pgid   uint64           /* pid of process group leader */
	pr_sid    uint64           /* session id */
	pr_ttydev uint64           /* controlling tty device */
	pr_addr   uint64           /* internal address of proc struct */
	pr_size   uint64           /* size of process image in KB (1024) units */
	pr_rssize uint64           /* resident set size in KB (1024) units */
	pr_start  pr_timestruc64_t /* process start time, time since epoch */
	pr_time   pr_timestruc64_t /* usr+sys cpu time for this process */
	pr_argv   uint64           /* address of initial argument vector in user process */
	pr_envp   uint64           /* address of initial environment vector in user process */
	pr_fname  [16]byte         /* last component of exec()ed pathname */
	pr_psargs [80]byte         /* initial characters of arg list */
	pr__pad   [8]uint64        /* reserved for future use */
	pr_lwp    lwpsInfo         /* "representative" thread info */
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
	rlimits, err := p.fillFromLimitsWithContext(ctx)
	if !gatherUsed || err != nil {
		return rlimits, err
	}

	_, _, _, _, rtprio, nice, _, err := p.fillFromStatWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := p.fillFromStatusWithContext(ctx); err != nil {
		return nil, err
	}

	for i := range rlimits {
		rs := &rlimits[i]
		switch rs.Resource {
		case RLIMIT_CPU:
			times, err := p.TimesWithContext(ctx)
			if err != nil {
				return nil, err
			}
			rs.Used = uint64(times.User + times.System)
		case RLIMIT_DATA:
			rs.Used = uint64(p.memInfo.Data)
		case RLIMIT_STACK:
			rs.Used = uint64(p.memInfo.Stack)
		case RLIMIT_RSS:
			rs.Used = uint64(p.memInfo.RSS)
		case RLIMIT_NOFILE:
			n, err := p.NumFDsWithContext(ctx)
			if err != nil {
				return nil, err
			}
			rs.Used = uint64(n)
		case RLIMIT_MEMLOCK:
			rs.Used = uint64(p.memInfo.Locked)
		case RLIMIT_AS:
			rs.Used = uint64(p.memInfo.VMS)
		case RLIMIT_LOCKS:
			// TODO we can get the used value from /proc/$pid/locks. But linux doesn't enforce it, so not a high priority.
		case RLIMIT_SIGPENDING:
			rs.Used = p.sigInfo.PendingProcess
		case RLIMIT_NICE:
			// The rlimit for nice is a little unusual, in that 0 means the niceness cannot be decreased beyond the current value, but it can be increased.
			// So effectively: if rs.Soft == 0 { rs.Soft = rs.Used }
			rs.Used = uint64(nice)
		case RLIMIT_RTPRIO:
			rs.Used = uint64(rtprio)
		}
	}

	return rlimits, err
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
	statFiles, err := filepath.Glob(common.HostProcWithContext(ctx, "[0-9]*/stat"))
	if err != nil {
		return nil, err
	}
	ret := make([]*Process, 0, len(statFiles))
	for _, statFile := range statFiles {
		statContents, err := os.ReadFile(statFile)
		if err != nil {
			continue
		}
		fields := splitProcStat(statContents)
		pid, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil {
			continue
		}
		ppid, err := strconv.ParseInt(fields[4], 10, 32)
		if err != nil {
			continue
		}
		if int32(ppid) == p.Pid {
			np, err := NewProcessWithContext(ctx, int32(pid))
			if err != nil {
				continue
			}
			ret = append(ret, np)
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Pid < ret[j].Pid })
	return ret, nil
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
	pid := p.Pid
	var ret []MemoryMapsStat
	smapsPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "smaps")
	if grouped {
		ret = make([]MemoryMapsStat, 1)
		// If smaps_rollup exists (require kernel >= 4.15), then we will use it
		// for pre-summed memory information for a process.
		smapsRollupPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "smaps_rollup")
		if _, err := os.Stat(smapsRollupPath); !os.IsNotExist(err) {
			smapsPath = smapsRollupPath
		}
	}
	contents, err := os.ReadFile(smapsPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")

	// function of parsing a block
	getBlock := func(firstLine []string, block []string) (MemoryMapsStat, error) {
		m := MemoryMapsStat{}
		if len(firstLine) >= 6 {
			m.Path = strings.Join(firstLine[5:], " ")
		}

		for _, line := range block {
			if strings.Contains(line, "VmFlags") {
				continue
			}
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			v := strings.Trim(field[1], "kB") // remove last "kB"
			v = strings.TrimSpace(v)
			t, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return m, err
			}

			switch field[0] {
			case "Size":
				m.Size = t
			case "Rss":
				m.Rss = t
			case "Pss":
				m.Pss = t
			case "Shared_Clean":
				m.SharedClean = t
			case "Shared_Dirty":
				m.SharedDirty = t
			case "Private_Clean":
				m.PrivateClean = t
			case "Private_Dirty":
				m.PrivateDirty = t
			case "Referenced":
				m.Referenced = t
			case "Anonymous":
				m.Anonymous = t
			case "Swap":
				m.Swap = t
			}
		}
		return m, nil
	}

	var firstLine []string
	blocks := make([]string, 0, 16)

	for i, line := range lines {
		fields := strings.Fields(line)
		if (len(fields) > 0 && !strings.HasSuffix(fields[0], ":")) || i == len(lines)-1 {
			// new block section
			if len(firstLine) > 0 && len(blocks) > 0 {
				g, err := getBlock(firstLine, blocks)
				if err != nil {
					return &ret, err
				}
				if grouped {
					ret[0].Size += g.Size
					ret[0].Rss += g.Rss
					ret[0].Pss += g.Pss
					ret[0].SharedClean += g.SharedClean
					ret[0].SharedDirty += g.SharedDirty
					ret[0].PrivateClean += g.PrivateClean
					ret[0].PrivateDirty += g.PrivateDirty
					ret[0].Referenced += g.Referenced
					ret[0].Anonymous += g.Anonymous
					ret[0].Swap += g.Swap
				} else {
					ret = append(ret, g)
				}
			}
			// starts new block
			blocks = make([]string, 0, 16)
			firstLine = fields
		} else {
			blocks = append(blocks, line)
		}
	}

	return &ret, nil
}

func (p *Process) EnvironWithContext(ctx context.Context) ([]string, error) {
	environPath := common.HostProcWithContext(ctx, strconv.Itoa(int(p.Pid)), "environ")

	environContent, err := os.ReadFile(environPath)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(environContent), "\000"), nil
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
	pid := p.Pid
	limitsFile := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "limits")
	d, err := os.Open(limitsFile)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	var limitStats []RlimitStat

	limitsScanner := bufio.NewScanner(d)
	for limitsScanner.Scan() {
		var statItem RlimitStat

		str := strings.Fields(limitsScanner.Text())

		// Remove the header line
		if strings.Contains(str[len(str)-1], "Units") {
			continue
		}

		// Assert that last item is a Hard limit
		statItem.Hard, err = limitToUint(str[len(str)-1])
		if err != nil {
			// On error remove last item and try once again since it can be unit or header line
			str = str[:len(str)-1]
			statItem.Hard, err = limitToUint(str[len(str)-1])
			if err != nil {
				return nil, err
			}
		}
		// Remove last item from string
		str = str[:len(str)-1]

		// Now last item is a Soft limit
		statItem.Soft, err = limitToUint(str[len(str)-1])
		if err != nil {
			return nil, err
		}
		// Remove last item from string
		str = str[:len(str)-1]

		// The rest is a stats name
		resourceName := strings.Join(str, " ")
		switch resourceName {
		case "Max cpu time":
			statItem.Resource = RLIMIT_CPU
		case "Max file size":
			statItem.Resource = RLIMIT_FSIZE
		case "Max data size":
			statItem.Resource = RLIMIT_DATA
		case "Max stack size":
			statItem.Resource = RLIMIT_STACK
		case "Max core file size":
			statItem.Resource = RLIMIT_CORE
		case "Max resident set":
			statItem.Resource = RLIMIT_RSS
		case "Max processes":
			statItem.Resource = RLIMIT_NPROC
		case "Max open files":
			statItem.Resource = RLIMIT_NOFILE
		case "Max locked memory":
			statItem.Resource = RLIMIT_MEMLOCK
		case "Max address space":
			statItem.Resource = RLIMIT_AS
		case "Max file locks":
			statItem.Resource = RLIMIT_LOCKS
		case "Max pending signals":
			statItem.Resource = RLIMIT_SIGPENDING
		case "Max msgqueue size":
			statItem.Resource = RLIMIT_MSGQUEUE
		case "Max nice priority":
			statItem.Resource = RLIMIT_NICE
		case "Max realtime priority":
			statItem.Resource = RLIMIT_RTPRIO
		case "Max realtime timeout":
			statItem.Resource = RLIMIT_RTTIME
		default:
			continue
		}

		limitStats = append(limitStats, statItem)
	}

	if err := limitsScanner.Err(); err != nil {
		return nil, err
	}

	return limitStats, nil
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
	pid := p.Pid
	exePath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "exe")
	exe, err := os.Readlink(exePath)
	if err != nil {
		return "", err
	}
	return string(exe), nil
}

// Get cmdline from /proc/(pid)/cmdline
func (p *Process) fillFromCmdlineWithContext(ctx context.Context) (string, error) {
	pid := p.Pid
	cmdPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := os.ReadFile(cmdPath)
	if err != nil {
		return "", err
	}
	ret := strings.FieldsFunc(string(cmdline), func(r rune) bool {
		return r == '\u0000'
	})

	return strings.Join(ret, " "), nil
}

func (p *Process) fillSliceFromCmdlineWithContext(ctx context.Context) ([]string, error) {
	pid := p.Pid
	cmdPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := os.ReadFile(cmdPath)
	if err != nil {
		return nil, err
	}
	if len(cmdline) == 0 {
		return nil, nil
	}

	cmdline = bytes.TrimRight(cmdline, "\x00")

	parts := bytes.Split(cmdline, []byte{0})
	var strParts []string
	for _, p := range parts {
		strParts = append(strParts, string(p))
	}

	return strParts, nil
}

// Get IO status from /proc/(pid)/io
func (p *Process) fillFromIOWithContext(ctx context.Context) (*IOCountersStat, error) {
	pid := p.Pid
	ioPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "io")
	ioline, err := os.ReadFile(ioPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(ioline), "\n")
	ret := &IOCountersStat{}

	for _, line := range lines {
		field := strings.Fields(line)
		if len(field) < 2 {
			continue
		}
		t, err := strconv.ParseUint(field[1], 10, 64)
		if err != nil {
			return nil, err
		}
		param := strings.TrimSuffix(field[0], ":")
		switch param {
		case "syscr":
			ret.ReadCount = t
		case "syscw":
			ret.WriteCount = t
		case "read_bytes":
			ret.DiskReadBytes = t
		case "write_bytes":
			ret.DiskWriteBytes = t
		case "rchar":
			ret.ReadBytes = t
		case "wchar":
			ret.WriteBytes = t
		}
	}

	return ret, nil
}

// Get memory info from /proc/(pid)/statm
func (p *Process) fillFromStatmWithContext(ctx context.Context) (*MemoryInfoStat, *MemoryInfoExStat, error) {
	pid := p.Pid
	memPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "statm")
	contents, err := os.ReadFile(memPath)
	if err != nil {
		return nil, nil, err
	}
	fields := strings.Split(string(contents), " ")

	vms, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	rss, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	memInfo := &MemoryInfoStat{
		RSS: rss * pageSize,
		VMS: vms * pageSize,
	}

	shared, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	text, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	lib, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	dirty, err := strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		return nil, nil, err
	}

	memInfoEx := &MemoryInfoExStat{
		RSS:    rss * pageSize,
		VMS:    vms * pageSize,
		Shared: shared * pageSize,
		Text:   text * pageSize,
		Lib:    lib * pageSize,
		Dirty:  dirty * pageSize,
	}

	return memInfo, memInfoEx, nil
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
	pid := p.Pid
	statPath := common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "comm")
	contents, err := os.ReadFile(statPath)
	if err != nil {
		return err
	}

	p.name = strings.TrimSuffix(string(contents), "\n")
	return nil
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
	var aixPSinfo aixPSInfo
	var aixlwpStat lwpStatus
	var aixlspPSinfo lwpsInfo
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

	ppid := aixStat.pr_ppid
	utime := float64(aixStat.pr_utime.tv_sec)
	stime := float64(aixStat.pr_stime.tv_sec)

	var iotime float64
	iotime = 0 // TODO: Figure out actual iotime for AIX

	cpuTimes := &cpu.TimesStat{
		CPU:    "cpu",
		User:   utime / float64(clockTicks),
		System: stime / float64(clockTicks),
		Iowait: iotime / float64(clockTicks),
	}

	bootTime, _ := common.BootTimeWithContext(ctx, invoke)
	startTime := uint64(aixPSinfo.pr_start.tv_sec)
	createTime := int64((startTime * 1000 / uint64(clockTicks)) + uint64(bootTime*1000))

	// This information is only available at thread level
	var rtpriority uint32
	var nice int32
	if tid > -1 {
		rtpriority = uint32(aixlspPSinfo.pr_pri)
		nice = int32(aixlspPSinfo.pr_nice)
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
