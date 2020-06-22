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
	// FieldChildren -- children is not pre-fetchable
	FieldBackground Field = iota + 1
	FieldCmdline
	FieldCmdlineSlice
	FieldConnections
	FieldCPUPercent
	FieldCreateTime
	FieldCwd
	FieldExe
	FieldForeground
	FieldGids
	FieldIOCounters
	FieldIOnice
	FieldIsRunning
	FieldMemoryInfo
	FieldMemoryInfoEx
	FieldMemoryMaps
	FieldMemoryMapsGrouped
	FieldMemoryPercent
	FieldName
	FieldNetIOCounters
	FieldNetIOCountersPerNic
	FieldNice
	FieldNumCtxSwitches
	FieldNumFDs
	FieldNumThreads
	FieldOpenFiles
	FieldPageFaults
	// FieldParent -- parent is not pre-fetchable
	FieldPpid
	FieldRlimit
	FieldRlimitUsage
	FieldStatus
	FieldTerminal
	FieldTgid
	FieldThreads
	FieldTimes
	FieldUids
	FieldUsername
)

func (f Field) String() string {
	switch f {
	case FieldBackground:
		return "Background"
	case FieldCmdline:
		return "Cmdline"
	case FieldCmdlineSlice:
		return "CmdlineSlice"
	case FieldConnections:
		return "Connections"
	case FieldCPUPercent:
		return "CPUPercent"
	case FieldCreateTime:
		return "CreateTime"
	case FieldCwd:
		return "Cwd"
	case FieldExe:
		return "Exe"
	case FieldForeground:
		return "Foreground"
	case FieldGids:
		return "Gids"
	case FieldIOCounters:
		return "IOCounters"
	case FieldIOnice:
		return "IOnice"
	case FieldIsRunning:
		return "IsRunning"
	case FieldMemoryInfo:
		return "MemoryInfo"
	case FieldMemoryInfoEx:
		return "MemoryInfoEx"
	case FieldMemoryMaps:
		return "MemoryMaps"
	case FieldMemoryMapsGrouped:
		return "MemoryMapsGrouped"
	case FieldMemoryPercent:
		return "MemoryPercent"
	case FieldName:
		return "Name"
	case FieldNetIOCounters:
		return "NetIOCounters"
	case FieldNetIOCountersPerNic:
		return "NetIOCountersPerNic"
	case FieldNice:
		return "Nice"
	case FieldNumCtxSwitches:
		return "NumCtxSwitches"
	case FieldNumFDs:
		return "NumFDs"
	case FieldNumThreads:
		return "NumThreads"
	case FieldOpenFiles:
		return "OpenFiles"
	case FieldPageFaults:
		return "PageFaults"
	case FieldPpid:
		return "Ppid"
	case FieldRlimit:
		return "Rlimit"
	case FieldRlimitUsage:
		return "RlimitUsage"
	case FieldStatus:
		return "Status"
	case FieldTerminal:
		return "Terminal"
	case FieldTgid:
		return "Tgid"
	case FieldThreads:
		return "Threads"
	case FieldTimes:
		return "Times"
	case FieldUids:
		return "Uids"
	case FieldUsername:
		return "Username"
	default:
		return "unknown"
	}
}

var AllFields = []Field{
	FieldBackground,
	FieldCmdline,
	FieldCmdlineSlice,
	FieldConnections,
	FieldCPUPercent,
	FieldCreateTime,
	FieldCwd,
	FieldExe,
	FieldForeground,
	FieldGids,
	FieldIOCounters,
	FieldIOnice,
	FieldIsRunning,
	FieldMemoryInfo,
	FieldMemoryInfoEx,
	FieldMemoryMaps,
	FieldMemoryMapsGrouped,
	FieldMemoryPercent,
	FieldName,
	FieldNetIOCounters,
	FieldNetIOCountersPerNic,
	FieldNice,
	FieldNumCtxSwitches,
	FieldNumFDs,
	FieldNumThreads,
	FieldOpenFiles,
	FieldPageFaults,
	FieldPpid,
	FieldRlimit,
	FieldRlimitUsage,
	FieldStatus,
	FieldTerminal,
	FieldTgid,
	FieldThreads,
	FieldTimes,
	FieldUids,
	FieldUsername,
}

type Process struct {
	Pid        int32 `json:"pid"`
	name       string
	createTime int64

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
func NewProcessWithFields(ctx context.Context, pid int32, fields ...Field) (*Process, error) {
	return newProcessWithFields(ctx, pid, nil, fields...)
}

func newProcessWithFields(ctx context.Context, pid int32, initialCache map[string]interface{}, fields ...Field) (*Process, error) {
	if initialCache == nil {
		initialCache = make(map[string]interface{})
	}

	p := &Process{
		Pid:   pid,
		cache: initialCache,
	}

	exists, err := PidExists(pid)
	if err != nil {
		return p, err
	}
	if !exists {
		return p, ErrorProcessNotRunning
	}

	requestedFields := make(map[Field]interface{}, len(fields))
	for _, f := range fields {
		requestedFields[f] = nil
	}

	err = p.prefetchFields(fields)
	if err != nil {
		return nil, err
	}

	p.requestedFields = requestedFields

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
	if !p.isFieldRequested(FieldBackground) {
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
	if !p.isFieldRequested(FieldIsRunning) {
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
	if !p.isFieldRequested(FieldCreateTime) {
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
	if !p.isFieldRequested(FieldMemoryPercent) {
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
	machineMemory, ok := p.cache["VirtualMemory"].(uint64)

	if !ok {
		tmp, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		machineMemory = tmp.Total
	}

	processMemory, err := p.MemoryInfo()
	if err != nil {
		return 0, err
	}
	used := processMemory.RSS

	return (100 * float32(used) / float32(machineMemory)), nil
}

// CPU_Percent returns how many percent of the CPU time this process uses
func (p *Process) CPUPercent() (float64, error) {
	return p.CPUPercentWithContext(context.Background())
}

func (p *Process) CPUPercentWithContext(ctx context.Context) (float64, error) {
	if !p.isFieldRequested(FieldCPUPercent) {
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

func (p *Process) genericPrefetchFields(ctx context.Context, fields []Field) {
	p.CreateTimeWithContext(ctx)

	for _, f := range fields {
		switch f {
		case FieldBackground:
			p.BackgroundWithContext(ctx)
		case FieldCmdline:
			p.CmdlineWithContext(ctx)
		case FieldCmdlineSlice:
			p.CmdlineSliceWithContext(ctx)
		case FieldConnections:
			p.ConnectionsWithContext(ctx)
		case FieldCPUPercent:
			p.CPUPercentWithContext(ctx)
		case FieldCreateTime:
			// already done
		case FieldCwd:
			p.CwdWithContext(ctx)
		case FieldExe:
			p.ExeWithContext(ctx)
		case FieldForeground:
			p.ForegroundWithContext(ctx)
		case FieldGids:
			p.GidsWithContext(ctx)
		case FieldIOCounters:
			p.IOCountersWithContext(ctx)
		case FieldIOnice:
			p.IOniceWithContext(ctx)
		case FieldIsRunning:
			p.IsRunningWithContext(ctx)
		case FieldMemoryInfo:
			p.MemoryInfoWithContext(ctx)
		case FieldMemoryInfoEx:
			p.MemoryInfoExWithContext(ctx)
		case FieldMemoryMaps:
			p.MemoryMapsWithContext(ctx, false)
		case FieldMemoryMapsGrouped:
			p.MemoryMapsWithContext(ctx, true)
		case FieldMemoryPercent:
			p.MemoryPercentWithContext(ctx)
		case FieldName:
			p.NameWithContext(ctx)
		case FieldNetIOCounters:
			p.NetIOCountersWithContext(ctx, false)
		case FieldNetIOCountersPerNic:
			p.NetIOCountersWithContext(ctx, true)
		case FieldNice:
			p.NiceWithContext(ctx)
		case FieldNumCtxSwitches:
			p.NumCtxSwitchesWithContext(ctx)
		case FieldNumFDs:
			p.NumFDsWithContext(ctx)
		case FieldNumThreads:
			p.NumThreadsWithContext(ctx)
		case FieldOpenFiles:
			p.OpenFilesWithContext(ctx)
		case FieldPageFaults:
			p.PageFaultsWithContext(ctx)
		case FieldPpid:
			p.PpidWithContext(ctx)
		case FieldRlimit:
			p.RlimitWithContext(ctx)
		case FieldRlimitUsage:
			p.RlimitUsageWithContext(ctx, true)
		case FieldStatus:
			p.StatusWithContext(ctx)
		case FieldTerminal:
			p.TerminalWithContext(ctx)
		case FieldTgid:
			p.Tgid()
		case FieldThreads:
			p.ThreadsWithContext(ctx)
		case FieldTimes:
			p.TimesWithContext(ctx)
		case FieldUids:
			p.UidsWithContext(ctx)
		case FieldUsername:
			p.UsernameWithContext(ctx)
		}
	}
}

func (p *Process) isFieldRequested(f Field) bool {
	if p.requestedFields == nil {
		return true
	}

	_, ok := p.requestedFields[f]

	return ok
}
