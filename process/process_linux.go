// +build linux

package process

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/internal/common"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/sys/unix"
)

var PageSize = uint64(os.Getpagesize())

const (
	PrioProcess = 0   // linux/resource.h
	ClockTicks  = 100 // C.sysconf(C._SC_CLK_TCK)
)

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

type statInfo struct {
	terminal   uint64
	ppid       int32
	cpuTimes   *cpu.TimesStat
	createTime int64
	rtpriority uint32
	nice       int32
	faults     *PageFaultsStat
	pgid       int32
	tpgid      int32
	err        error
}

type statusInfo struct {
	numCtxSwitches *NumCtxSwitchesStat
	memInfo        *MemoryInfoStat
	sigInfo        *SignalInfoStat
	name           string
	status         string
	parent         int32
	tgid           int32
	uids           []int32
	gids           []int32
	numThreads     int32
	err            error
}

type fdInfo struct {
	numFDs    int32
	openfiles []*OpenFilesStat
	err       error
}

type fdListInfo struct {
	statPath string
	fnames   []string
	err      error
}

type statmInfo struct {
	meminfo   *MemoryInfoStat
	meminfoEx *MemoryInfoExStat
	err       error
}

// Ppid returns Parent Process ID of the process.
func (p *Process) Ppid() (int32, error) {
	return p.PpidWithContext(context.Background())
}

func (p *Process) PpidWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(Ppid) {
		return -1, ErrorFieldNotRequested
	}

	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return -1, ret.err
	}
	return ret.ppid, nil
}

// Name returns name of the process.
func (p *Process) Name() (string, error) {
	return p.NameWithContext(context.Background())
}

func (p *Process) NameWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Name) {
		return "", ErrorFieldNotRequested
	}

	if p.name == "" {
		ret := p.fillFromStatusWithContext(ctx)
		if ret.err != nil {
			return "", ret.err
		}

		p.name = ret.name
	}
	return p.name, nil
}

// Tgid returns tgid, a Linux-synonym for user-space Pid
func (p *Process) Tgid() (int32, error) {
	if !p.isFieldRequested(Tgid) {
		return 0, ErrorFieldNotRequested
	}

	if p.tgid == 0 {
		ret := p.fillFromStatusWithContext(context.Background())
		if ret.err != nil {
			return 0, ret.err
		}

		p.tgid = ret.tgid
	}
	return p.tgid, nil
}

// Exe returns executable path of the process.
func (p *Process) Exe() (string, error) {
	return p.ExeWithContext(context.Background())
}

func (p *Process) ExeWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Exe) {
		return "", ErrorFieldNotRequested
	}

	return p.fillFromExeWithContext(ctx)
}

// Cmdline returns the command line arguments of the process as a string with
// each argument separated by 0x20 ascii character.
func (p *Process) Cmdline() (string, error) {
	return p.CmdlineWithContext(context.Background())
}

func (p *Process) CmdlineWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Cmdline) {
		return "", ErrorFieldNotRequested
	}

	return p.fillFromCmdlineWithContext(ctx)
}

// CmdlineSlice returns the command line arguments of the process as a slice with each
// element being an argument.
func (p *Process) CmdlineSlice() ([]string, error) {
	return p.CmdlineSliceWithContext(context.Background())
}

func (p *Process) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	if !p.isFieldRequested(CmdlineSlice) {
		return nil, ErrorFieldNotRequested
	}

	return p.fillSliceFromCmdlineWithContext(ctx)
}

func (p *Process) createTimeWithContext(ctx context.Context) (int64, error) {
	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return 0, ret.err
	}
	return ret.createTime, nil
}

// Cwd returns current working directory of the process.
func (p *Process) Cwd() (string, error) {
	return p.CwdWithContext(context.Background())
}

func (p *Process) CwdWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Cwd) {
		return "", ErrorFieldNotRequested
	}

	return p.fillFromCwdWithContext(ctx)
}

// Parent returns parent Process of the process.
func (p *Process) Parent() (*Process, error) {
	return p.ParentWithContext(context.Background())
}

func (p *Process) ParentWithContext(ctx context.Context) (*Process, error) {
	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return nil, ret.err
	}
	if ret.parent == 0 {
		return nil, fmt.Errorf("wrong number of parents")
	}

	fields := make([]Field, 0, len(p.requestedFields))
	for f := range p.requestedFields {
		fields = append(fields, f)
	}

	if p.requestedFields != nil {
		return NewProcessWithFields(ret.parent, fields...)
	}

	return NewProcess(ret.parent)
}

// Status returns the process status.
// Return value could be one of these.
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
// The character is same within all supported platforms.
func (p *Process) Status() (string, error) {
	return p.StatusWithContext(context.Background())
}

func (p *Process) StatusWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Status) {
		return "", ErrorFieldNotRequested
	}

	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return "", ret.err
	}
	return ret.status, nil
}

// Foreground returns true if the process is in foreground, false otherwise.
func (p *Process) Foreground() (bool, error) {
	return p.ForegroundWithContext(context.Background())
}

func (p *Process) ForegroundWithContext(ctx context.Context) (bool, error) {
	if !p.isFieldRequested(Foreground) {
		return false, ErrorFieldNotRequested
	}

	return p.foregroundWithContext(ctx)
}

func (p *Process) foregroundWithContext(ctx context.Context) (bool, error) {
	// see https://github.com/shirou/gopsutil/issues/596#issuecomment-432707831 for implementation details
	ret := p.fillFromStatWithContext(ctx)
	return ret.pgid == ret.tpgid, ret.err
}

// Uids returns user ids of the process as a slice of the int
func (p *Process) Uids() ([]int32, error) {
	return p.UidsWithContext(context.Background())
}

func (p *Process) UidsWithContext(ctx context.Context) ([]int32, error) {
	if !p.isFieldRequested(Uids) {
		return nil, ErrorFieldNotRequested
	}

	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return []int32{}, ret.err
	}
	return ret.uids, nil
}

// Gids returns group ids of the process as a slice of the int
func (p *Process) Gids() ([]int32, error) {
	return p.GidsWithContext(context.Background())
}

func (p *Process) GidsWithContext(ctx context.Context) ([]int32, error) {
	if !p.isFieldRequested(Gids) {
		return nil, ErrorFieldNotRequested
	}

	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return []int32{}, ret.err
	}
	return ret.gids, nil
}

// Terminal returns a terminal which is associated with the process.
func (p *Process) Terminal() (string, error) {
	return p.TerminalWithContext(context.Background())
}

func (p *Process) TerminalWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(Terminal) {
		return "", ErrorFieldNotRequested
	}

	cacheKey := "Terminal"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.terminalWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(string), v.err
}

func (p *Process) terminalWithContextNoCache(ctx context.Context) (string, error) {
	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return "", ret.err
	}
	termmap, err := getTerminalMap()
	if err != nil {
		return "", err
	}
	terminal := termmap[ret.terminal]
	return terminal, nil
}

// Nice returns a nice value (priority).
// Notice: gopsutil can not set nice value.
func (p *Process) Nice() (int32, error) {
	return p.NiceWithContext(context.Background())
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(Nice) {
		return 0, ErrorFieldNotRequested
	}

	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return 0, ret.err
	}
	return ret.nice, nil
}

// IOnice returns process I/O nice value (priority).
func (p *Process) IOnice() (int32, error) {
	return p.IOniceWithContext(context.Background())
}

func (p *Process) IOniceWithContext(ctx context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

// Rlimit returns Resource Limits.
func (p *Process) Rlimit() ([]RlimitStat, error) {
	return p.RlimitWithContext(context.Background())
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	return p.RlimitUsage(false)
}

// RlimitUsage returns Resource Limits.
// If gatherUsed is true, the currently used value will be gathered and added
// to the resulting RlimitStat.
func (p *Process) RlimitUsage(gatherUsed bool) ([]RlimitStat, error) {
	return p.RlimitUsageWithContext(context.Background(), gatherUsed)
}

func (p *Process) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	f := Rlimit
	if gatherUsed {
		f = RlimitUsage
	}

	if !p.isFieldRequested(f) {
		return nil, ErrorFieldNotRequested
	}

	rlimits, err := p.fillFromLimitsWithContext(ctx)
	if !gatherUsed || err != nil {
		return rlimits, err
	}

	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return nil, ret.err
	}
	status := p.fillFromStatusWithContext(ctx)
	if status.err != nil {
		return nil, status.err
	}

	// Copy the rlimits objet to avoid modifying cache value that may be used for
	// Rlimit / RlimitUsage gatherUsed=false
	if p.cache != nil {
		tmp := make([]RlimitStat, len(rlimits))
		for i, x := range rlimits {
			tmp[i] = x
		}

		rlimits = tmp
	}

	for i := range rlimits {
		rs := &rlimits[i]
		switch rs.Resource {
		case RLIMIT_CPU:
			times, err := p.timesWithContext(ctx)
			if err != nil {
				return nil, err
			}
			rs.Used = uint64(times.User + times.System)
		case RLIMIT_DATA:
			rs.Used = uint64(status.memInfo.Data)
		case RLIMIT_STACK:
			rs.Used = uint64(status.memInfo.Stack)
		case RLIMIT_RSS:
			rs.Used = uint64(status.memInfo.RSS)
		case RLIMIT_NOFILE:
			n, err := p.numFDsWithContext(ctx)
			if err != nil {
				return nil, err
			}
			rs.Used = uint64(n)
		case RLIMIT_MEMLOCK:
			rs.Used = uint64(status.memInfo.Locked)
		case RLIMIT_AS:
			rs.Used = uint64(status.memInfo.VMS)
		case RLIMIT_LOCKS:
			//TODO we can get the used value from /proc/$pid/locks. But linux doesn't enforce it, so not a high priority.
		case RLIMIT_SIGPENDING:
			rs.Used = status.sigInfo.PendingProcess
		case RLIMIT_NICE:
			// The rlimit for nice is a little unusual, in that 0 means the niceness cannot be decreased beyond the current value, but it can be increased.
			// So effectively: if rs.Soft == 0 { rs.Soft = rs.Used }
			rs.Used = uint64(ret.nice)
		case RLIMIT_RTPRIO:
			rs.Used = uint64(ret.rtpriority)
		}
	}

	return rlimits, err
}

// IOCounters returns IO Counters.
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return p.IOCountersWithContext(context.Background())
}

func (p *Process) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	if !p.isFieldRequested(IOCounters) {
		return nil, ErrorFieldNotRequested
	}

	return p.fillFromIOWithContext(ctx)
}

// NumCtxSwitches returns the number of the context switches of the process.
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return p.NumCtxSwitchesWithContext(context.Background())
}

func (p *Process) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	if !p.isFieldRequested(NumCtxSwitches) {
		return nil, ErrorFieldNotRequested
	}

	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return nil, ret.err
	}
	return ret.numCtxSwitches, nil
}

// NumFDs returns the number of File Descriptors used by the process.
func (p *Process) NumFDs() (int32, error) {
	return p.NumFDsWithContext(context.Background())
}

func (p *Process) NumFDsWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(NumFDs) {
		return 0, ErrorFieldNotRequested
	}
	return p.numFDsWithContext(ctx)
}

func (p *Process) numFDsWithContext(ctx context.Context) (int32, error) {
	_, fnames, err := p.fillFromfdListWithContext(ctx)
	return int32(len(fnames)), err
}

// NumThreads returns the number of threads used by the process.
func (p *Process) NumThreads() (int32, error) {
	return p.NumThreadsWithContext(context.Background())
}

func (p *Process) NumThreadsWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(NumThreads) {
		return 0, ErrorFieldNotRequested
	}

	ret := p.fillFromStatusWithContext(ctx)
	if ret.err != nil {
		return 0, ret.err
	}
	return ret.numThreads, nil
}

func (p *Process) Threads() (map[int32]*cpu.TimesStat, error) {
	return p.ThreadsWithContext(context.Background())
}

func (p *Process) ThreadsWithContext(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	if !p.isFieldRequested(Threads) {
		return nil, ErrorFieldNotRequested
	}

	cacheKey := "Threads"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.threadsWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(map[int32]*cpu.TimesStat), v.err
}

func (p *Process) threadsWithContextNoCache(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	ret := make(map[int32]*cpu.TimesStat)
	taskPath := common.HostProc(strconv.Itoa(int(p.Pid)), "task")

	tids, err := readPidsFromDir(taskPath)
	if err != nil {
		return nil, err
	}

	for _, tid := range tids {
		tmp := p.fillFromTIDStatWithContext(ctx, tid)
		if tmp.err != nil {
			return nil, tmp.err
		}
		ret[tid] = tmp.cpuTimes
	}

	return ret, nil
}

// Times returns CPU times of the process.
func (p *Process) Times() (*cpu.TimesStat, error) {
	return p.TimesWithContext(context.Background())
}

func (p *Process) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	if !p.isFieldRequested(Times) {
		return nil, ErrorFieldNotRequested
	}
	return p.timesWithContext(ctx)
}

func (p *Process) timesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return nil, ret.err
	}
	return ret.cpuTimes, nil
}

// CPUAffinity returns CPU affinity of the process.
//
// Notice: Not implemented yet.
func (p *Process) CPUAffinity() ([]int32, error) {
	return p.CPUAffinityWithContext(context.Background())
}

func (p *Process) CPUAffinityWithContext(ctx context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
}

// MemoryInfo returns platform in-dependend memory information, such as RSS, VMS and Swap
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	return p.MemoryInfoWithContext(context.Background())
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {
	if !p.isFieldRequested(MemoryInfo) {
		return nil, ErrorFieldNotRequested
	}

	meminfo, _, err := p.fillFromStatmWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return meminfo, nil
}

// MemoryInfoEx returns platform dependend memory information.
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return p.MemoryInfoExWithContext(context.Background())
}

func (p *Process) MemoryInfoExWithContext(ctx context.Context) (*MemoryInfoExStat, error) {
	if !p.isFieldRequested(MemoryInfoEx) {
		return nil, ErrorFieldNotRequested
	}

	_, memInfoEx, err := p.fillFromStatmWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return memInfoEx, nil
}

// PageFaultsInfo returns the process's page fault counters
func (p *Process) PageFaults() (*PageFaultsStat, error) {
	return p.PageFaultsWithContext(context.Background())
}

func (p *Process) PageFaultsWithContext(ctx context.Context) (*PageFaultsStat, error) {
	if !p.isFieldRequested(PageFaults) {
		return nil, ErrorFieldNotRequested
	}

	ret := p.fillFromStatWithContext(ctx)
	if ret.err != nil {
		return nil, ret.err
	}
	return ret.faults, nil

}

// Children returns a slice of Process of the process.
func (p *Process) Children() ([]*Process, error) {
	return p.ChildrenWithContext(context.Background())
}

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
	pids, err := common.CallPgrepWithContext(ctx, invoke, p.Pid)
	if err != nil {
		if pids == nil || len(pids) == 0 {
			return nil, ErrorNoChildren
		}
		return nil, err
	}

	fields := make([]Field, 0, len(p.requestedFields))
	for f := range p.requestedFields {
		fields = append(fields, f)
	}

	ret := make([]*Process, 0, len(pids))
	for _, pid := range pids {
		var (
			np  *Process
			err error
		)

		if p.requestedFields != nil {
			np, err = NewProcessWithFields(pid, fields...)
		} else {
			np, err = NewProcess(pid)
		}
		if err != nil {
			return nil, err
		}
		ret = append(ret, np)
	}
	return ret, nil
}

// OpenFiles returns a slice of OpenFilesStat opend by the process.
// OpenFilesStat includes a file path and file descriptor.
func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return p.OpenFilesWithContext(context.Background())
}

func (p *Process) OpenFilesWithContext(ctx context.Context) ([]OpenFilesStat, error) {
	if !p.isFieldRequested(OpenFiles) {
		return nil, ErrorFieldNotRequested
	}

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

// Connections returns a slice of net.ConnectionStat used by the process.
// This returns all kind of the connection. This measn TCP, UDP or UNIX.
func (p *Process) Connections() ([]net.ConnectionStat, error) {
	return p.ConnectionsWithContext(context.Background())
}

func (p *Process) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	if !p.isFieldRequested(Connections) {
		return nil, ErrorFieldNotRequested
	}

	cacheKey := "Connections"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := net.ConnectionsPid("all", p.Pid)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.([]net.ConnectionStat), v.err
}

// Connections returns a slice of net.ConnectionStat used by the process at most `max`
func (p *Process) ConnectionsMax(max int) ([]net.ConnectionStat, error) {
	return p.ConnectionsMaxWithContext(context.Background(), max)
}

func (p *Process) ConnectionsMaxWithContext(ctx context.Context, max int) ([]net.ConnectionStat, error) {
	if p.requestedFields != nil {
		return nil, ErrorFieldNotRequested
	}

	return net.ConnectionsPidMax("all", p.Pid, max)
}

// NetIOCounters returns NetIOCounters of the process.
func (p *Process) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	return p.NetIOCountersWithContext(context.Background(), pernic)
}

func (p *Process) NetIOCountersWithContext(ctx context.Context, pernic bool) ([]net.IOCountersStat, error) {
	field := NetIOCounters
	cacheKey := "NetIOCounters"
	if pernic {
		field = NetIOCountersPerNic
		cacheKey = "NetIOCountersPerNic"
	}

	if !p.isFieldRequested(field) {
		return nil, ErrorFieldNotRequested
	}

	// Ideally we could derive NetIOCounters from NetIOCountersPerNic,
	// but it require to duplicate the private method net.getIOCountersAll
	v, ok := p.cache[cacheKey].(valueOrError)
	if !ok {
		filename := common.HostProc(strconv.Itoa(int(p.Pid)), "net/dev")
		tmp, err := net.IOCountersByFile(pernic, filename)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.([]net.IOCountersStat), v.err
}

// MemoryMaps get memory maps from /proc/(pid)/smaps
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	return p.MemoryMapsWithContext(context.Background(), grouped)
}

func (p *Process) MemoryMapsWithContext(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	field := MemoryMaps
	cacheKey := "MemoryMaps"
	if grouped {
		cacheKey = "MemoryMapsGrouped"
		field = MemoryMapsGrouped
	}

	if !p.isFieldRequested(field) {
		return nil, ErrorFieldNotRequested
	}

	v, ok := p.cache[cacheKey].(valueOrError)

	// We could derive MemoryMapsGrouped from MemoryMaps. But that means user
	// that only care about MemoryMapsGrouped will pay the memory to hold
	// full MemoryMaps.
	if !ok {
		tmp, err := p.memoryMapsWithContextNoCache(ctx, grouped)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(*[]MemoryMapsStat), v.err
}

func (p *Process) memoryMapsWithContextNoCache(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	pid := p.Pid
	var ret []MemoryMapsStat
	if grouped {
		ret = make([]MemoryMapsStat, 1)
	}
	smapsPath := common.HostProc(strconv.Itoa(int(pid)), "smaps")
	contents, err := ioutil.ReadFile(smapsPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")

	// function of parsing a block
	getBlock := func(first_line []string, block []string) (MemoryMapsStat, error) {
		m := MemoryMapsStat{}
		m.Path = first_line[len(first_line)-1]

		for _, line := range block {
			if strings.Contains(line, "VmFlags") {
				continue
			}
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			v := strings.Trim(field[1], " kB") // remove last "kB"
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

	blocks := make([]string, 16)
	for _, line := range lines {
		// This is a quick fix for #879
		line = strings.ReplaceAll(line, "\t", " ")
		field := strings.Split(line, " ")
		if strings.HasSuffix(field[0], ":") == false {
			// new block section
			if len(blocks) > 0 {
				g, err := getBlock(field, blocks)
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
			blocks = make([]string, 16)
		} else {
			blocks = append(blocks, line)
		}
	}

	return &ret, nil
}

/**
** Internal functions
**/

func limitToInt(val string) (int32, error) {
	if val == "unlimited" {
		return math.MaxInt32, nil
	} else {
		res, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(res), nil
	}
}

// Get num_fds from /proc/(pid)/limits
func (p *Process) fillFromLimitsWithContext(ctx context.Context) ([]RlimitStat, error) {
	cacheKey := "fillFromLimits"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.fillFromLimitsWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.([]RlimitStat), v.err
}

func (p *Process) fillFromLimitsWithContextNoCache(ctx context.Context) ([]RlimitStat, error) {
	pid := p.Pid
	limitsFile := common.HostProc(strconv.Itoa(int(pid)), "limits")
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
		statItem.Hard, err = limitToInt(str[len(str)-1])
		if err != nil {
			// On error remove last item an try once again since it can be unit or header line
			str = str[:len(str)-1]
			statItem.Hard, err = limitToInt(str[len(str)-1])
			if err != nil {
				return nil, err
			}
		}
		// Remove last item from string
		str = str[:len(str)-1]

		//Now last item is a Soft limit
		statItem.Soft, err = limitToInt(str[len(str)-1])
		if err != nil {
			return nil, err
		}
		// Remove last item from string
		str = str[:len(str)-1]

		//The rest is a stats name
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
	cacheKey := "fillFromfdList"
	v, ok := p.cache[cacheKey].(fdListInfo)

	if !ok {
		statPath, fnames, err := p.fillFromfdListWithContextNoCache(ctx)
		v = fdListInfo{
			statPath: statPath,
			fnames:   fnames,
			err:      err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.statPath, v.fnames, v.err
}

func (p *Process) fillFromfdListWithContextNoCache(ctx context.Context) (string, []string, error) {
	pid := p.Pid
	statPath := common.HostProc(strconv.Itoa(int(pid)), "fd")
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
	cacheKey := "fillFromfdList"
	v, ok := p.cache[cacheKey].(fdInfo)

	if !ok {
		numFDs, openfiles, err := p.fillFromfdWithContextNoCache(ctx)
		v = fdInfo{
			numFDs:    numFDs,
			openfiles: openfiles,
			err:       err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.numFDs, v.openfiles, v.err
}

func (p *Process) fillFromfdWithContextNoCache(ctx context.Context) (int32, []*OpenFilesStat, error) {
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
	cacheKey := "fillFromCwd"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.fillFromCwdWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(string), v.err
}

func (p *Process) fillFromCwdWithContextNoCache(ctx context.Context) (string, error) {
	pid := p.Pid
	cwdPath := common.HostProc(strconv.Itoa(int(pid)), "cwd")
	cwd, err := os.Readlink(cwdPath)
	if err != nil {
		return "", err
	}
	return string(cwd), nil
}

// Get exe from /proc/(pid)/exe
func (p *Process) fillFromExeWithContext(ctx context.Context) (string, error) {
	cacheKey := "fillFromExe"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.fillFromExeWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(string), v.err
}

func (p *Process) fillFromExeWithContextNoCache(ctx context.Context) (string, error) {
	pid := p.Pid
	exePath := common.HostProc(strconv.Itoa(int(pid)), "exe")
	exe, err := os.Readlink(exePath)
	if err != nil {
		return "", err
	}
	return string(exe), nil
}

// Get cmdline from /proc/(pid)/cmdline
func (p *Process) fillFromCmdlineWithContext(ctx context.Context) (string, error) {
	ret, err := p.fillSliceFromCmdlineWithContext(ctx)
	return strings.Join(ret, " "), err
}

func (p *Process) fillSliceFromCmdlineWithContext(ctx context.Context) ([]string, error) {
	cacheKey := "fillSliceFromCmdline"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.fillSliceFromCmdlineWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.([]string), v.err
}

func (p *Process) fillSliceFromCmdlineWithContextNoCache(ctx context.Context) ([]string, error) {
	pid := p.Pid
	cmdPath := common.HostProc(strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return nil, err
	}
	if len(cmdline) == 0 {
		return nil, nil
	}
	if cmdline[len(cmdline)-1] == 0 {
		cmdline = cmdline[:len(cmdline)-1]
	}
	parts := bytes.Split(cmdline, []byte{0})
	var strParts []string
	for _, p := range parts {
		strParts = append(strParts, string(p))
	}

	return strParts, nil
}

// Get IO status from /proc/(pid)/io
func (p *Process) fillFromIOWithContext(ctx context.Context) (*IOCountersStat, error) {
	cacheKey := "fillFromIO"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.fillFromIOWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(*IOCountersStat), v.err
}

func (p *Process) fillFromIOWithContextNoCache(ctx context.Context) (*IOCountersStat, error) {
	pid := p.Pid
	ioPath := common.HostProc(strconv.Itoa(int(pid)), "io")
	ioline, err := ioutil.ReadFile(ioPath)
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
		param := field[0]
		if strings.HasSuffix(param, ":") {
			param = param[:len(param)-1]
		}
		switch param {
		case "syscr":
			ret.ReadCount = t
		case "syscw":
			ret.WriteCount = t
		case "read_bytes":
			ret.ReadBytes = t
		case "write_bytes":
			ret.WriteBytes = t
		}
	}

	return ret, nil
}

// Get memory info from /proc/(pid)/statm
func (p *Process) fillFromStatmWithContext(ctx context.Context) (*MemoryInfoStat, *MemoryInfoExStat, error) {
	cacheKey := "fillFromStatm"
	v, ok := p.cache[cacheKey].(statmInfo)

	if !ok {
		meminfo, meminfoEx, err := p.fillFromStatmWithContextNoCache(ctx)
		v = statmInfo{
			meminfo:   meminfo,
			meminfoEx: meminfoEx,
			err:       err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.meminfo, v.meminfoEx, v.err
}

func (p *Process) fillFromStatmWithContextNoCache(ctx context.Context) (*MemoryInfoStat, *MemoryInfoExStat, error) {
	pid := p.Pid
	memPath := common.HostProc(strconv.Itoa(int(pid)), "statm")
	contents, err := ioutil.ReadFile(memPath)
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
		RSS: rss * PageSize,
		VMS: vms * PageSize,
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
		RSS:    rss * PageSize,
		VMS:    vms * PageSize,
		Shared: shared * PageSize,
		Text:   text * PageSize,
		Lib:    lib * PageSize,
		Dirty:  dirty * PageSize,
	}

	return memInfo, memInfoEx, nil
}

// Get various status from /proc/(pid)/status
func (p *Process) fillFromStatusWithContext(ctx context.Context) statusInfo {
	cacheKey := "fillFromStatus"
	v, ok := p.cache[cacheKey].(statusInfo)

	if !ok {
		v = p.fillFromStatusWithContextNoCache(ctx)
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v
}

func (p *Process) fillFromStatusWithContextNoCache(ctx context.Context) statusInfo {
	result := statusInfo{}

	pid := p.Pid
	statPath := common.HostProc(strconv.Itoa(int(pid)), "status")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		result.err = err
		return result
	}
	lines := strings.Split(string(contents), "\n")
	result.numCtxSwitches = &NumCtxSwitchesStat{}
	result.memInfo = &MemoryInfoStat{}
	result.sigInfo = &SignalInfoStat{}
	for _, line := range lines {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		value := tabParts[1]
		switch strings.TrimRight(tabParts[0], ":") {
		case "Name":
			result.name = strings.Trim(value, " \t")
			if len(result.name) >= 15 {
				cmdlineSlice, err := p.CmdlineSlice()
				if err != nil {
					result.err = err
					return result
				}
				if len(cmdlineSlice) > 0 {
					extendedName := filepath.Base(cmdlineSlice[0])
					if strings.HasPrefix(extendedName, result.name) {
						result.name = extendedName
					} else {
						result.name = cmdlineSlice[0]
					}
				}
			}
		case "State":
			result.status = value[0:1]
		case "PPid", "Ppid":
			pval, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				result.err = err
				return result
			}
			result.parent = int32(pval)
		case "Tgid":
			pval, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				result.err = err
				return result
			}
			result.tgid = int32(pval)
		case "Uid":
			result.uids = make([]int32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					result.err = err
					return result
				}
				result.uids = append(result.uids, int32(v))
			}
		case "Gid":
			result.gids = make([]int32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					result.err = err
					return result
				}
				result.gids = append(result.gids, int32(v))
			}
		case "Threads":
			v, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				result.err = err
				return result
			}
			result.numThreads = int32(v)
		case "voluntary_ctxt_switches":
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.numCtxSwitches.Voluntary = v
		case "nonvoluntary_ctxt_switches":
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.numCtxSwitches.Involuntary = v
		case "VmRSS":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.RSS = v * 1024
		case "VmSize":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.VMS = v * 1024
		case "VmSwap":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.Swap = v * 1024
		case "VmHWM":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.HWM = v * 1024
		case "VmData":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.Data = v * 1024
		case "VmStk":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.Stack = v * 1024
		case "VmLck":
			value := strings.Trim(value, " kB") // remove last "kB"
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.memInfo.Locked = v * 1024
		case "SigPnd":
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.sigInfo.PendingThread = v
		case "ShdPnd":
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.sigInfo.PendingProcess = v
		case "SigBlk":
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.sigInfo.Blocked = v
		case "SigIgn":
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.sigInfo.Ignored = v
		case "SigCgt":
			v, err := strconv.ParseUint(value, 16, 64)
			if err != nil {
				result.err = err
				return result
			}
			result.sigInfo.Caught = v
		}

	}
	return result
}

func (p *Process) fillFromTIDStatWithContext(ctx context.Context, tid int32) statInfo {
	if tid != -1 {
		// don't cache tid != -1 here, it's cached in ThreadsWithContext
		return p.fillFromTIDStatWithContextNoCache(ctx, tid)
	}

	cacheKey := "fillFromTIDStat"
	v, ok := p.cache[cacheKey].(statInfo)

	if !ok {
		v = p.fillFromTIDStatWithContextNoCache(ctx, tid)
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v
}

func (p *Process) fillFromTIDStatWithContextNoCache(ctx context.Context, tid int32) statInfo {
	pid := p.Pid
	var statPath string

	if tid == -1 {
		statPath = common.HostProc(strconv.Itoa(int(pid)), "stat")
	} else {
		statPath = common.HostProc(strconv.Itoa(int(pid)), "task", strconv.Itoa(int(tid)), "stat")
	}

	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return statInfo{err: err}
	}
	fields := strings.Fields(string(contents))

	i := 1
	for !strings.HasSuffix(fields[i], ")") {
		i++
	}

	terminal, err := strconv.ParseUint(fields[i+5], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}

	ppid, err := strconv.ParseInt(fields[i+2], 10, 32)
	if err != nil {
		return statInfo{err: err}
	}
	utime, err := strconv.ParseFloat(fields[i+12], 64)
	if err != nil {
		return statInfo{err: err}
	}

	stime, err := strconv.ParseFloat(fields[i+13], 64)
	if err != nil {
		return statInfo{err: err}
	}

	// There is no such thing as iotime in stat file.  As an approximation, we
	// will use delayacct_blkio_ticks (aggregated block I/O delays, as per Linux
	// docs).  Note: I am assuming at least Linux 2.6.18
	iotime, err := strconv.ParseFloat(fields[i+40], 64)
	if err != nil {
		iotime = 0 // Ancient linux version, most likely
	}

	cpuTimes := &cpu.TimesStat{
		CPU:    "cpu",
		User:   float64(utime / ClockTicks),
		System: float64(stime / ClockTicks),
		Iowait: float64(iotime / ClockTicks),
	}

	bootTime, _ := common.BootTimeWithContext(ctx)
	t, err := strconv.ParseUint(fields[i+20], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}
	ctime := (t / uint64(ClockTicks)) + uint64(bootTime)
	createTime := int64(ctime * 1000)

	rtpriority, err := strconv.ParseInt(fields[i+16], 10, 32)
	if err != nil {
		return statInfo{err: err}
	}
	if rtpriority < 0 {
		rtpriority = rtpriority*-1 - 1
	} else {
		rtpriority = 0
	}

	//	p.Nice = mustParseInt32(fields[18])
	// use syscall instead of parse Stat file
	snice, _ := unix.Getpriority(PrioProcess, int(pid))
	nice := int32(snice) // FIXME: is this true?

	minFault, err := strconv.ParseUint(fields[i+8], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}
	cMinFault, err := strconv.ParseUint(fields[i+9], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}
	majFault, err := strconv.ParseUint(fields[i+10], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}
	cMajFault, err := strconv.ParseUint(fields[i+11], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}

	faults := &PageFaultsStat{
		MinorFaults:      minFault,
		MajorFaults:      majFault,
		ChildMinorFaults: cMinFault,
		ChildMajorFaults: cMajFault,
	}

	pgid, err := strconv.ParseInt(fields[i+3], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}

	tpgid, err := strconv.ParseInt(fields[i+6], 10, 64)
	if err != nil {
		return statInfo{err: err}
	}

	return statInfo{
		terminal:   terminal,
		ppid:       int32(ppid),
		cpuTimes:   cpuTimes,
		createTime: createTime,
		rtpriority: uint32(rtpriority),
		nice:       nice,
		faults:     faults,
		pgid:       int32(pgid),
		tpgid:      int32(tpgid),
	}
}

func (p *Process) fillFromStatWithContext(ctx context.Context) statInfo {
	return p.fillFromTIDStatWithContext(ctx, -1)
}

func pidsWithContext(ctx context.Context) ([]int32, error) {
	return readPidsFromDir(common.HostProc())
}

// Process returns a slice of pointers to Process structs for all
// currently running processes.
func Processes() ([]*Process, error) {
	return ProcessesWithContext(context.Background())
}

func ProcessesWithContext(ctx context.Context) ([]*Process, error) {
	out := []*Process{}

	pids, err := PidsWithContext(ctx)
	if err != nil {
		return out, err
	}

	for _, pid := range pids {
		p, err := NewProcess(pid)
		if err != nil {
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

func ProcessesWithFields(ctx context.Context, fields ...Field) ([]*Process, error) {
	out := []*Process{}

	pids, err := PidsWithContext(ctx)
	if err != nil {
		return out, err
	}

	machineMemory := uint64(0)
	for _, f := range fields {
		if f == MemoryPercent {
			tmp, err := mem.VirtualMemory()
			if err == nil {
				machineMemory = tmp.Total
			}
		}
	}

	for _, pid := range pids {
		p, err := newProcessWithFields(pid, machineMemory, fields...)
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
