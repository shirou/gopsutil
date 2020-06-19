package process

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/internal/common"
	"github.com/shirou/gopsutil/mem"
)

var (
	invoke                 common.Invoker = common.Invoke{}
	ErrorNoChildren                       = errors.New("process does not have children")
	ErrorProcessNotRunning                = errors.New("process does not exist")
	ErrorFieldNotRequested                = errors.New("this fields was not requested")
	ErrorRequireUpdate                    = errors.New("this method require updates and cannot be used with NewProcessWithFields")
)

type Field uint

const (
	// Children -- children is not pre-fetchable
	Background Field = iota + 1
	Cmdline
	CmdlineSlice
	Connections
	CPUPercent
	CreateTime
	Cwd
	Exe
	Foreground
	Gids
	IOCounters
	IOnice
	IsRunning
	MemoryInfo
	MemoryInfoEx
	MemoryMaps
	MemoryMapsGrouped
	MemoryPercent
	Name
	NetIOCounters
	NetIOCountersPerNic
	Nice
	NumCtxSwitches
	NumFDs
	NumThreads
	OpenFiles
	PageFaults
	// Parent -- parent is not pre-fetchable
	Ppid
	Rlimit
	RlimitUsage
	Status
	Terminal
	Tgid
	Threads
	Times
	Uids
	Username
)

func (f Field) String() string {
	switch f {
	case Background:
		return "Background"
	case Cmdline:
		return "Cmdline"
	case CmdlineSlice:
		return "CmdlineSlice"
	case Connections:
		return "Connections"
	case CPUPercent:
		return "CPUPercent"
	case CreateTime:
		return "CreateTime"
	case Cwd:
		return "Cwd"
	case Exe:
		return "Exe"
	case Foreground:
		return "Foreground"
	case Gids:
		return "Gids"
	case IOCounters:
		return "IOCounters"
	case IOnice:
		return "IOnice"
	case IsRunning:
		return "IsRunning"
	case MemoryInfo:
		return "MemoryInfo"
	case MemoryInfoEx:
		return "MemoryInfoEx"
	case MemoryMaps:
		return "MemoryMaps"
	case MemoryMapsGrouped:
		return "MemoryMapsGrouped"
	case MemoryPercent:
		return "MemoryPercent"
	case Name:
		return "Name"
	case NetIOCounters:
		return "NetIOCounters"
	case NetIOCountersPerNic:
		return "NetIOCountersPerNic"
	case Nice:
		return "Nice"
	case NumCtxSwitches:
		return "NumCtxSwitches"
	case NumFDs:
		return "NumFDs"
	case NumThreads:
		return "NumThreads"
	case OpenFiles:
		return "OpenFiles"
	case PageFaults:
		return "PageFaults"
	case Ppid:
		return "Ppid"
	case Rlimit:
		return "Rlimit"
	case RlimitUsage:
		return "RlimitUsage"
	case Status:
		return "Status"
	case Terminal:
		return "Terminal"
	case Tgid:
		return "Tgid"
	case Threads:
		return "Threads"
	case Times:
		return "Times"
	case Uids:
		return "Uids"
	case Username:
		return "Username"
	default:
		return "unknown"
	}
}

var AllFields = []Field{
	Background,
	Cmdline,
	CmdlineSlice,
	Connections,
	CPUPercent,
	CreateTime,
	Cwd,
	Exe,
	Foreground,
	Gids,
	IOCounters,
	IOnice,
	IsRunning,
	MemoryInfo,
	MemoryInfoEx,
	MemoryMaps,
	MemoryMapsGrouped,
	MemoryPercent,
	Name,
	NetIOCounters,
	NetIOCountersPerNic,
	Nice,
	NumCtxSwitches,
	NumFDs,
	NumThreads,
	OpenFiles,
	PageFaults,
	Ppid,
	Rlimit,
	RlimitUsage,
	Status,
	Terminal,
	Tgid,
	Threads,
	Times,
	Uids,
	Username,
}

type Process struct {
	Pid           int32 `json:"pid"`
	name          string
	createTime    int64
	machineMemory uint64

	cache           map[string]interface{}
	requestedFields map[Field]interface{}

	lastCPUTimes *cpu.TimesStat
	lastCPUTime  time.Time

	tgid int32
}

type valueOrError struct {
	value interface{}
	err   error
}

type OpenFilesStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

type MemoryInfoStat struct {
	RSS    uint64 `json:"rss"`    // bytes
	VMS    uint64 `json:"vms"`    // bytes
	HWM    uint64 `json:"hwm"`    // bytes
	Data   uint64 `json:"data"`   // bytes
	Stack  uint64 `json:"stack"`  // bytes
	Locked uint64 `json:"locked"` // bytes
	Swap   uint64 `json:"swap"`   // bytes
}

type SignalInfoStat struct {
	PendingProcess uint64 `json:"pending_process"`
	PendingThread  uint64 `json:"pending_thread"`
	Blocked        uint64 `json:"blocked"`
	Ignored        uint64 `json:"ignored"`
	Caught         uint64 `json:"caught"`
}

type RlimitStat struct {
	Resource int32  `json:"resource"`
	Soft     int32  `json:"soft"` //TODO too small. needs to be uint64
	Hard     int32  `json:"hard"` //TODO too small. needs to be uint64
	Used     uint64 `json:"used"`
}

type IOCountersStat struct {
	ReadCount  uint64 `json:"readCount"`
	WriteCount uint64 `json:"writeCount"`
	ReadBytes  uint64 `json:"readBytes"`
	WriteBytes uint64 `json:"writeBytes"`
}

type NumCtxSwitchesStat struct {
	Voluntary   int64 `json:"voluntary"`
	Involuntary int64 `json:"involuntary"`
}

type PageFaultsStat struct {
	MinorFaults      uint64 `json:"minorFaults"`
	MajorFaults      uint64 `json:"majorFaults"`
	ChildMinorFaults uint64 `json:"childMinorFaults"`
	ChildMajorFaults uint64 `json:"childMajorFaults"`
}

// Resource limit constants are from /usr/include/x86_64-linux-gnu/bits/resource.h
// from libc6-dev package in Ubuntu 16.10
const (
	RLIMIT_CPU        int32 = 0
	RLIMIT_FSIZE      int32 = 1
	RLIMIT_DATA       int32 = 2
	RLIMIT_STACK      int32 = 3
	RLIMIT_CORE       int32 = 4
	RLIMIT_RSS        int32 = 5
	RLIMIT_NPROC      int32 = 6
	RLIMIT_NOFILE     int32 = 7
	RLIMIT_MEMLOCK    int32 = 8
	RLIMIT_AS         int32 = 9
	RLIMIT_LOCKS      int32 = 10
	RLIMIT_SIGPENDING int32 = 11
	RLIMIT_MSGQUEUE   int32 = 12
	RLIMIT_NICE       int32 = 13
	RLIMIT_RTPRIO     int32 = 14
	RLIMIT_RTTIME     int32 = 15
)

func (p Process) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func (o OpenFilesStat) String() string {
	s, _ := json.Marshal(o)
	return string(s)
}

func (m MemoryInfoStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (r RlimitStat) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}

func (i IOCountersStat) String() string {
	s, _ := json.Marshal(i)
	return string(s)
}

func (p NumCtxSwitchesStat) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

// Pids returns a slice of process ID list which are running now.
func Pids() ([]int32, error) {
	return PidsWithContext(context.Background())
}

func PidsWithContext(ctx context.Context) ([]int32, error) {
	pids, err := pidsWithContext(ctx)
	sort.Slice(pids, func(i, j int) bool { return pids[i] < pids[j] })
	return pids, err
}

// NewProcess creates a new Process instance, it only stores the pid and
// checks that the process exists. Other method on Process can be used
// to get more information about the process. An error will be returned
// if the process does not exist.
func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	exists, err := PidExists(pid)
	if err != nil {
		return p, err
	}
	if !exists {
		return p, ErrorProcessNotRunning
	}
	p.CreateTime()
	return p, nil
}

// NewProcessWithFields creates a new Process instance, and pre-fetch all requested
// fields. Those fields and only those fields will be available and won't be updated
// after creation.
// This method may be faster if you retrive multiple fields, because unlike NewProcess where
// files are parsed for each method call, here files are parsed once at creation.
func NewProcessWithFields(pid int32, fields ...Field) (*Process, error) {
	return newProcessWithFields(pid, 0, fields...)
}

func newProcessWithFields(pid int32, machineMemory uint64, fields ...Field) (*Process, error) {
	p := &Process{
		Pid:           pid,
		cache:         make(map[string]interface{}),
		machineMemory: machineMemory,
	}

	exists, err := PidExists(pid)
	if err != nil {
		return p, err
	}
	if !exists {
		return p, ErrorProcessNotRunning
	}

	p.prefetchFields(fields)

	return p, nil
}

func PidExists(pid int32) (bool, error) {
	return PidExistsWithContext(context.Background(), pid)
}

// Background returns true if the process is in background, false otherwise.
func (p *Process) Background() (bool, error) {
	return p.BackgroundWithContext(context.Background())
}

func (p *Process) BackgroundWithContext(ctx context.Context) (bool, error) {
	if !p.isFieldRequested(Background) {
		return false, ErrorFieldNotRequested
	}

	fg, err := p.foregroundWithContext(ctx)
	if err != nil {
		return false, err
	}
	return !fg, err
}

// If interval is 0, return difference from last call(non-blocking).
// If interval > 0, wait interval sec and return diffrence between start and end.
func (p *Process) Percent(interval time.Duration) (float64, error) {
	return p.PercentWithContext(context.Background(), interval)
}

func (p *Process) PercentWithContext(ctx context.Context, interval time.Duration) (float64, error) {
	if p.cache != nil {
		return 0, ErrorRequireUpdate
	}

	cpuTimes, err := p.Times()
	if err != nil {
		return 0, err
	}
	now := time.Now()

	if interval > 0 {
		p.lastCPUTimes = cpuTimes
		p.lastCPUTime = now
		time.Sleep(interval)
		cpuTimes, err = p.Times()
		now = time.Now()
		if err != nil {
			return 0, err
		}
	} else {
		if p.lastCPUTimes == nil {
			// invoked first time
			p.lastCPUTimes = cpuTimes
			p.lastCPUTime = now
			return 0, nil
		}
	}

	numcpu := runtime.NumCPU()
	delta := (now.Sub(p.lastCPUTime).Seconds()) * float64(numcpu)
	ret := calculatePercent(p.lastCPUTimes, cpuTimes, delta, numcpu)
	p.lastCPUTimes = cpuTimes
	p.lastCPUTime = now
	return ret, nil
}

// IsRunning returns whether the process is still running or not.
func (p *Process) IsRunning() (bool, error) {
	return p.IsRunningWithContext(context.Background())
}

func (p *Process) IsRunningWithContext(ctx context.Context) (bool, error) {
	if !p.isFieldRequested(IsRunning) {
		return false, ErrorFieldNotRequested
	}

	cacheKey := "IsRunning"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.isRunningWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(bool), v.err
}

func (p *Process) isRunningWithContextNoCache(ctx context.Context) (bool, error) {
	createTime, err := p.CreateTimeWithContext(ctx)
	if err != nil {
		return false, err
	}
	p2, err := NewProcess(p.Pid)
	if err == ErrorProcessNotRunning {
		return false, nil
	}
	createTime2, err := p2.CreateTimeWithContext(ctx)
	if err != nil {
		return false, err
	}
	return createTime == createTime2, nil
}

// CreateTime returns created time of the process in milliseconds since the epoch, in UTC.
func (p *Process) CreateTime() (int64, error) {
	return p.CreateTimeWithContext(context.Background())
}

func (p *Process) CreateTimeWithContext(ctx context.Context) (int64, error) {
	if !p.isFieldRequested(CreateTime) {
		return 0, ErrorFieldNotRequested
	}

	if p.createTime != 0 {
		return p.createTime, nil
	}
	createTime, err := p.createTimeWithContext(ctx)
	p.createTime = createTime
	return p.createTime, err
}

func calculatePercent(t1, t2 *cpu.TimesStat, delta float64, numcpu int) float64 {
	if delta == 0 {
		return 0
	}
	delta_proc := t2.Total() - t1.Total()
	overall_percent := ((delta_proc / delta) * 100) * float64(numcpu)
	return overall_percent
}

// MemoryPercent returns how many percent of the total RAM this process uses
func (p *Process) MemoryPercent() (float32, error) {
	return p.MemoryPercentWithContext(context.Background())
}

func (p *Process) MemoryPercentWithContext(ctx context.Context) (float32, error) {
	if !p.isFieldRequested(MemoryPercent) {
		return 0, ErrorFieldNotRequested
	}

	cacheKey := "MemoryPercent"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.memoryPercentWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(float32), v.err
}

func (p *Process) memoryPercentWithContextNoCache(ctx context.Context) (float32, error) {
	if p.machineMemory == 0 || p.cache == nil {
		machineMemory, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		p.machineMemory = machineMemory.Total
	}

	processMemory, err := p.MemoryInfo()
	if err != nil {
		return 0, err
	}
	used := processMemory.RSS

	return (100 * float32(used) / float32(p.machineMemory)), nil
}

// CPU_Percent returns how many percent of the CPU time this process uses
func (p *Process) CPUPercent() (float64, error) {
	return p.CPUPercentWithContext(context.Background())
}

func (p *Process) CPUPercentWithContext(ctx context.Context) (float64, error) {
	if !p.isFieldRequested(CPUPercent) {
		return 0, ErrorFieldNotRequested
	}

	cacheKey := "CPUPercent"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.cpuPercentWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(float64), v.err
}

func (p *Process) cpuPercentWithContextNoCache(ctx context.Context) (float64, error) {
	crt_time, err := p.CreateTime()
	if err != nil {
		return 0, err
	}

	cput, err := p.Times()
	if err != nil {
		return 0, err
	}

	created := time.Unix(0, crt_time*int64(time.Millisecond))
	totalTime := time.Since(created).Seconds()
	if totalTime <= 0 {
		return 0, nil
	}

	return 100 * cput.Total() / totalTime, nil
}

func (p *Process) prefetchFields(fields []Field) {
	ctx := context.Background()

	requestedFields := make(map[Field]interface{}, len(fields))

	p.CreateTimeWithContext(ctx)

	for _, f := range fields {
		requestedFields[f] = nil

		switch f {
		case Background:
			p.BackgroundWithContext(ctx)
		case Cmdline:
			p.CmdlineWithContext(ctx)
		case CmdlineSlice:
			p.CmdlineSliceWithContext(ctx)
		case Connections:
			p.ConnectionsWithContext(ctx)
		case CPUPercent:
			p.CPUPercentWithContext(ctx)
		case CreateTime:
			// already done
		case Cwd:
			p.CwdWithContext(ctx)
		case Exe:
			p.ExeWithContext(ctx)
		case Foreground:
			p.ForegroundWithContext(ctx)
		case Gids:
			p.GidsWithContext(ctx)
		case IOCounters:
			p.IOCountersWithContext(ctx)
		case IOnice:
			p.IOniceWithContext(ctx)
		case IsRunning:
			p.IsRunningWithContext(ctx)
		case MemoryInfo:
			p.MemoryInfoWithContext(ctx)
		case MemoryInfoEx:
			p.MemoryInfoExWithContext(ctx)
		case MemoryMaps:
			p.MemoryMapsWithContext(ctx, false)
		case MemoryMapsGrouped:
			p.MemoryMapsWithContext(ctx, true)
		case MemoryPercent:
			p.MemoryPercentWithContext(ctx)
		case Name:
			p.NameWithContext(ctx)
		case NetIOCounters:
			p.NetIOCountersWithContext(ctx, false)
		case NetIOCountersPerNic:
			p.NetIOCountersWithContext(ctx, true)
		case Nice:
			p.NiceWithContext(ctx)
		case NumCtxSwitches:
			p.NumCtxSwitchesWithContext(ctx)
		case NumFDs:
			p.NumFDsWithContext(ctx)
		case NumThreads:
			p.NumThreadsWithContext(ctx)
		case OpenFiles:
			p.OpenFilesWithContext(ctx)
		case PageFaults:
			p.PageFaultsWithContext(ctx)
		case Ppid:
			p.PpidWithContext(ctx)
		case Rlimit:
			p.RlimitWithContext(ctx)
		case RlimitUsage:
			p.RlimitUsageWithContext(ctx, true)
		case Status:
			p.StatusWithContext(ctx)
		case Terminal:
			p.TerminalWithContext(ctx)
		case Tgid:
			p.Tgid()
		case Threads:
			p.ThreadsWithContext(ctx)
		case Times:
			p.TimesWithContext(ctx)
		case Uids:
			p.UidsWithContext(ctx)
		case Username:
			p.UsernameWithContext(ctx)
		}
	}

	p.prefetchFieldsPlatformSpecific(fields)

	p.requestedFields = requestedFields
}

func (p *Process) isFieldRequested(f Field) bool {
	if p.requestedFields == nil {
		return true
	}

	_, ok := p.requestedFields[f]

	return ok
}
