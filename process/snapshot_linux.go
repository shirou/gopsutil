// +build linux

package process

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// Function are keys indicating the functions that should be collected in a Snapshot
type Function uint

const (
	Ppid Function = iota
	PpidWithContext
	Name
	NameWithContext
	Tgid
	Exe
	ExeWithContext
	Cmdline
	CmdlineWithContext
	CmdlineSlice
	CmdlineSliceWithContext
	CreateTime
	CreateTimeWithContext
	Cwd
	CwdWithContext
	Parent
	ParentWithContext
	Status
	StatusWithContext
	Uids
	UidsWithContext
	Gids
	GidsWithContext
	Terminal
	TerminalWithContext
	Nice
	NiceWithContext
	IOnice
	IOniceWithContext
	Rlimit
	RlimitWithContext
	RlimitUsage
	RlimitUsageWithContext
	IOCounters
	IOCountersWithContext
	NumCtxSwitches
	NumCtxSwitchesWithContext
	NumFDs
	NumFDsWithContext
	NumThreads
	NumThreadsWithContext
	Threads
	ThreadsWithContext
	Times
	TimesWithContext
	CPUAffinity
	CPUAffinityWithContext
	MemoryInfo
	MemoryInfoWithContext
	MemoryInfoEx
	MemoryInfoExWithContext
	Children
	ChildrenWithContext
	OpenFiles
	OpenFilesWithContext
	Connections
	ConnectionsWithContext
	NetIOCounters
	NetIOCountersWithContext
	IsRunning
	IsRunningWithContext
	MemoryMaps
	MemoryMapsWithContext
	CPUPercent
	CPUPercentWithContext
	MemoryPercent
	MemoryPercentWithContext
)

// internalFunctionKey is a key used to refer to the functions in the internalFunctions array
type internalFunctionKey uint

// these keys are used to ensure order of execution.
// the order here is the order in which the functions will be executed
// a function with dependencies should below the dependant function
const (
	fillFromCmdlineWithContext internalFunctionKey = iota
	fillFromStatWithContext
	fillSliceFromCmdlineWithContext
	fillFromCwdWithContext
	fillFromExeWithContext
	fillFromfdWithContext
	fillFromIOWithContext
	fillFromStatmWithContext
	fillFromStatusWithContext
	fillThreads
	fillMemoryMaps
	fillFromLimitsWithContext
	getRlimitUsage
	getTerminalFromMap
	fillConnections
	fillNetIOCounters
	fillCPUPercent
	fillMemoryPercent
)

// internalFunctions is an array of functions that invoke fill functions or perform 1 time calculations
var internalFunctions = []func(s *Snapshot, ctx context.Context) (err error){
	fillFromStatWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.t, s.parent, s.lastCPUTimes, s.createTime, s.rtprio, s.nice, err = s.fillFromStatWithContext(ctx)
		return
	},
	fillSliceFromCmdlineWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.cmdLineSlice, err = s.fillSliceFromCmdlineWithContext(ctx)
		return
	},
	fillFromCmdlineWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.cmdLine, err = s.fillFromCmdlineWithContext(ctx)
		return
	},
	fillFromCwdWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.cwd, err = s.fillFromCwdWithContext(ctx)
		return
	},
	fillFromExeWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.exe, err = s.fillFromExeWithContext(ctx)
		return
	},
	fillFromfdWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		if s.numFds, s.openFiles, err = s.fillFromfdWithContext(ctx); err != nil {
			return
		}
		return
	},
	fillFromIOWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.iOCounters, err = s.fillFromIOWithContext(ctx)
		return
	},
	fillFromStatmWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.memInfo, s.memInfoEx, err = s.fillFromStatmWithContext(ctx)
		return
	},
	fillFromStatusWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		err = s.fillFromStatusWithContext(ctx)
		return
	},
	fillThreads: func(s *Snapshot, ctx context.Context) (err error) {
		s.threads, err = s.Process.ThreadsWithContext(ctx)
		return
	},
	fillMemoryMaps: func(s *Snapshot, ctx context.Context) (err error) {
		if s.memoryMaps, err = s.Process.MemoryMapsWithContext(ctx, false); err != nil {
			return
		}
		s.memoryMapsGrouped, err = s.Process.MemoryMapsWithContext(ctx, true)
		return
	},
	fillFromLimitsWithContext: func(s *Snapshot, ctx context.Context) (err error) {
		s.rlimits, err = s.fillFromLimitsWithContext(ctx)
		return
	},
	getRlimitUsage: func(s *Snapshot, ctx context.Context) (err error) {
		// depends: fillFromStatWithContext, fillFromLimitsWithContext
		s.rlimitsUsage, err = s.getRlimitUsage(s.rlimits, s.rtprio, s.nice, s.lastCPUTimes)
		return
	},
	getTerminalFromMap: func(s *Snapshot, ctx context.Context) (err error) {
		// depends: fillFromStatWithContext
		s.terminal, err = s.getTerminalFromMap(s.t)
		return
	},
	fillConnections: func(s *Snapshot, ctx context.Context) (err error) {
		s.connections, err = s.Process.ConnectionsWithContext(ctx)
		return
	},
	fillNetIOCounters: func(s *Snapshot, ctx context.Context) (err error) {
		if s.netIOCounters, err = s.Process.NetIOCountersWithContext(ctx, false); err != nil {
			return
		}
		s.netIOCountersPerNic, err = s.Process.NetIOCountersWithContext(ctx, true)
		return
	},
	fillCPUPercent: func(s *Snapshot, ctx context.Context) (err error) {
		// depends: on fillFromStatWithContext
		s.cpuPercent, err = s.calculateCPUPercent(ctx, s.createTime, s.lastCPUTimes)
		return
	},
	fillMemoryPercent: func(s *Snapshot, ctx context.Context) (err error) {
		// depends: on fillFromStatmWithContext
		machineMemory, err := mem.VirtualMemory()
		if err != nil {
			return
		}
		s.memPercent = (100 * float32(s.memInfo.RSS) / float32(machineMemory.Total))
		return
	},
}

// map internalFunctionKey functions to external functions
var functionToInternal = map[Function]map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){
	Ppid:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	PpidWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	Name:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	NameWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Tgid:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Exe:                     map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromExeWithContext: internalFunctions[fillFromExeWithContext]},
	ExeWithContext:          map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromExeWithContext: internalFunctions[fillFromExeWithContext]},
	Cmdline:                 map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromCmdlineWithContext: internalFunctions[fillFromCmdlineWithContext]},
	CmdlineWithContext:      map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromCmdlineWithContext: internalFunctions[fillFromCmdlineWithContext]},
	CmdlineSlice:            map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillSliceFromCmdlineWithContext: internalFunctions[fillSliceFromCmdlineWithContext]},
	CmdlineSliceWithContext: map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillSliceFromCmdlineWithContext: internalFunctions[fillSliceFromCmdlineWithContext]},
	CreateTime:              map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	CreateTimeWithContext:   map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	Cwd:                       map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromCwdWithContext: internalFunctions[fillFromCwdWithContext]},
	CwdWithContext:            map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromCwdWithContext: internalFunctions[fillFromCwdWithContext]},
	Parent:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	ParentWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	Status:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	StatusWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Uids:                      map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	UidsWithContext:           map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Gids:                      map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	GidsWithContext:           map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Terminal:                  map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], getTerminalFromMap: internalFunctions[getTerminalFromMap]},
	TerminalWithContext:       map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], getTerminalFromMap: internalFunctions[getTerminalFromMap]},
	Nice:                      map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	NiceWithContext:           map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	IOnice:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	IOniceWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	Rlimit:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromLimitsWithContext: internalFunctions[fillFromLimitsWithContext]},
	RlimitWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromLimitsWithContext: internalFunctions[fillFromLimitsWithContext]},
	RlimitUsage:               map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], fillFromLimitsWithContext: internalFunctions[fillFromLimitsWithContext], getRlimitUsage: internalFunctions[getRlimitUsage]},
	RlimitUsageWithContext:    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], fillFromLimitsWithContext: internalFunctions[fillFromLimitsWithContext], getRlimitUsage: internalFunctions[getRlimitUsage]},
	IOCounters:                map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromIOWithContext: internalFunctions[fillFromIOWithContext]},
	IOCountersWithContext:     map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromIOWithContext: internalFunctions[fillFromIOWithContext]},
	NumCtxSwitches:            map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	NumCtxSwitchesWithContext: map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	NumFDs:                   map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromfdWithContext: internalFunctions[fillFromfdWithContext]},
	NumFDsWithContext:        map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromfdWithContext: internalFunctions[fillFromfdWithContext]},
	NumThreads:               map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	NumThreadsWithContext:    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatusWithContext: internalFunctions[fillFromStatusWithContext]},
	Threads:                  map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillThreads: internalFunctions[fillThreads]},
	ThreadsWithContext:       map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillThreads: internalFunctions[fillThreads]},
	Times:                    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	TimesWithContext:         map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext]},
	CPUAffinity:              map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	CPUAffinityWithContext:   map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	MemoryInfo:               map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext]},
	MemoryInfoWithContext:    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext]},
	MemoryInfoEx:             map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext]},
	MemoryInfoExWithContext:  map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext]},
	Children:                 map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	ChildrenWithContext:      map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	OpenFiles:                map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromfdWithContext: internalFunctions[fillFromfdWithContext]},
	OpenFilesWithContext:     map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromfdWithContext: internalFunctions[fillFromfdWithContext]},
	Connections:              map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillConnections: internalFunctions[fillConnections]},
	ConnectionsWithContext:   map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillConnections: internalFunctions[fillConnections]},
	NetIOCounters:            map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillNetIOCounters: internalFunctions[fillNetIOCounters]},
	NetIOCountersWithContext: map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillNetIOCounters: internalFunctions[fillNetIOCounters]},
	IsRunning:                map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	IsRunningWithContext:     map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){},
	MemoryMaps:               map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillMemoryMaps: internalFunctions[fillMemoryMaps]},
	MemoryMapsWithContext:    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillMemoryMaps: internalFunctions[fillMemoryMaps]},
	CPUPercent:               map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], fillCPUPercent: internalFunctions[fillCPUPercent]},
	CPUPercentWithContext:    map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatWithContext: internalFunctions[fillFromStatWithContext], fillCPUPercent: internalFunctions[fillCPUPercent]},
	MemoryPercent:            map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext], fillMemoryPercent: internalFunctions[fillMemoryPercent]},
	MemoryPercentWithContext: map[internalFunctionKey]func(s *Snapshot, ctx context.Context) (err error){fillFromStatmWithContext: internalFunctions[fillFromStatmWithContext], fillMemoryPercent: internalFunctions[fillMemoryPercent]},
}

// Snapshot is process information at a moment in time
type Snapshot struct {
	*CommonSnapshot
	createTime          int64
	rtprio              uint32
	nice                int32
	cmdLine             string
	cmdLineSlice        []string
	connections         []net.ConnectionStat
	cpuPercent          float64
	cwd                 string
	exe                 string
	iOCounters          *IOCountersStat
	memInfo             *MemoryInfoStat
	memInfoEx           *MemoryInfoExStat
	memoryInfoEx        *MemoryInfoExStat
	memoryMaps          *[]MemoryMapsStat
	memoryMapsGrouped   *[]MemoryMapsStat
	memPercent          float32
	netIOCounters       []net.IOCountersStat
	netIOCountersPerNic []net.IOCountersStat
	numFds              int32
	openFiles           []*OpenFilesStat
	threads             map[int32]*cpu.TimesStat
	rlimits             []RlimitStat
	rlimitsUsage        []RlimitStat
	terminal            string
	t                   uint64
}

// Ppid returns Parent Process ID of the process.
func (s *Snapshot) Ppid() (int32, error) {
	return s.parent, nil
}

// PpidWithContext
func (s *Snapshot) PpidWithContext(ctx context.Context) (int32, error) {
	return s.parent, nil
}

// Name returns name of the process.
func (s *Snapshot) Name() (string, error) {
	return s.name, nil
}

// NameWithContext returns name of the process.
func (s *Snapshot) NameWithContext(ctx context.Context) (string, error) {
	return s.name, nil
}

// Tgid returns tgid, a Linux-synonym for user-space Pid //TODO: set this?
func (s *Snapshot) Tgid() (int32, error) {
	return s.tgid, nil
}

// Exe returns executable path of the process.
func (s *Snapshot) Exe() (string, error) {
	return s.exe, nil
}

// ExeWithContext returns executable path of the process.
func (s *Snapshot) ExeWithContext(ctx context.Context) (string, error) {
	return s.exe, nil
}

// Cmdline returns the command line arguments of the process as a string with
// each argument separated by 0x20 ascii character.
func (s *Snapshot) Cmdline() (string, error) {
	return s.cmdLine, nil
}

// CmdlineWithContext returns the command line arguments of the process as a string with
// each argument separated by 0x20 ascii character.
func (s *Snapshot) CmdlineWithContext(ctx context.Context) (string, error) {
	return s.cmdLine, nil
}

// CmdlineSlice returns the command line arguments of the process as a slice with each
// element being an argument.
func (s *Snapshot) CmdlineSlice() ([]string, error) {
	return s.cmdLineSlice, nil
}

// CmdlineSliceWithContext returns the command line arguments of the process as a slice with each
// element being an argument.
func (s *Snapshot) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	return s.cmdLineSlice, nil
}

// CreateTime returns created time of the process in milliseconds since the epoch, in UTC.
func (s *Snapshot) CreateTime() (int64, error) {
	return s.createTime, nil
}

// CreateTimeWithContext returns created time of the process in milliseconds since the epoch, in UTC.
func (s *Snapshot) CreateTimeWithContext(ctx context.Context) (int64, error) {
	return s.createTime, nil
}

// Cwd returns current working directory of the process.
func (s *Snapshot) Cwd() (string, error) {
	return s.cwd, nil
}

// CwdWithContext returns current working directory of the process.
func (s *Snapshot) CwdWithContext(ctx context.Context) (string, error) {
	return s.cwd, nil
}

// Parent returns parent Process of the process.
func (s *Snapshot) Parent() (*Process, error) {
	return s.ParentWithContext(context.Background())
}

// ParentWithContext returns parent Process of the process.
func (s *Snapshot) ParentWithContext(ctx context.Context) (*Process, error) {
	if s.parent == 0 {
		return nil, fmt.Errorf("wrong number of parents")
	}
	return NewProcess(s.parent)
}

// Status returns the process status.
// Return value could be one of these.
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
// The charactor is same within all supported platforms.
func (s *Snapshot) Status() (string, error) {
	return s.status, nil
}

// StatusWithContext returns the process status.
// Return value could be one of these.
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
// The charactor is same within all supported platforms.
func (s *Snapshot) StatusWithContext(ctx context.Context) (string, error) {
	return s.status, nil
}

// Uids returns user ids of the process as a slice of the int
func (s *Snapshot) Uids() ([]int32, error) {
	return s.uids, nil
}

// UidsWithContext returns user ids of the process as a slice of the int
func (s *Snapshot) UidsWithContext(ctx context.Context) ([]int32, error) {
	return s.uids, nil
}

// Gids returns group ids of the process as a slice of the int
func (s *Snapshot) Gids() ([]int32, error) {
	return s.gids, nil
}

// GidsWithContext returns group ids of the process as a slice of the int
func (s *Snapshot) GidsWithContext(ctx context.Context) ([]int32, error) {
	return s.gids, nil
}

// Terminal returns a terminal which is associated with the process.
func (s *Snapshot) Terminal() (string, error) {
	return s.terminal, nil
}

// TerminalWithContext returns a terminal which is associated with the process.
func (s *Snapshot) TerminalWithContext(ctx context.Context) (string, error) {
	return s.terminal, nil
}

// Nice returns a nice value (priority).
// Notice: gopsutil can not set nice value.
func (s *Snapshot) Nice() (int32, error) {
	return s.nice, nil
}

// NiceWithContext returns a nice value (priority).
// Notice: gopsutil can not set nice value.
func (s *Snapshot) NiceWithContext(ctx context.Context) (int32, error) {
	return s.nice, nil
}

// TODO: add IOnice once it's implemented

// Rlimit returns Resource Limits.
func (s *Snapshot) Rlimit() ([]RlimitStat, error) {
	return s.rlimits, nil
}

// Rlimit returns Resource Limits.
func (s *Snapshot) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	return s.rlimits, nil
}

// RlimitUsage returns Resource Limits.
// If gatherUsed is true, the currently used value will be gathered and added
// to the resulting RlimitStat.
func (s *Snapshot) RlimitUsage(gatherUsed bool) ([]RlimitStat, error) {
	return s.rlimitsUsage, nil
}

// RlimitUsageWithContext returns Resource Limits.
// If gatherUsed is true, the currently used value will be gathered and added
// to the resulting RlimitStat.
func (s *Snapshot) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	return s.rlimitsUsage, nil
}

// IOCounters returns IO Counters.
func (s *Snapshot) IOCounters() (*IOCountersStat, error) {
	return s.iOCounters, nil
}

// IOCountersWithContext returns IO Counters.
func (s *Snapshot) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	return s.iOCounters, nil
}

// NumCtxSwitches returns the number of the context switches of the process.
func (s *Snapshot) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return s.numCtxSwitches, nil
}

// NumCtxSwitchesWithContext returns the number of the context switches of the process.
func (s *Snapshot) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	return s.numCtxSwitches, nil
}

// NumFDs returns the number of File Descriptors used by the process.
func (s *Snapshot) NumFDs() (int32, error) {
	return s.numFds, nil
}

// NumFDsWithContext returns the number of File Descriptors used by the process.
func (s *Snapshot) NumFDsWithContext(ctx context.Context) (int32, error) {
	return s.numFds, nil
}

// NumThreads returns the number of threads used by the process.
func (s *Snapshot) NumThreads() (int32, error) {
	return s.numThreads, nil
}

// NumThreadsWithContext returns the number of threads used by the process.
func (s *Snapshot) NumThreadsWithContext(ctx context.Context) (int32, error) {
	return s.numThreads, nil
}

// Threads returns the cpu.TimesStat for each thread
func (s *Snapshot) Threads() (map[int32]*cpu.TimesStat, error) {
	return s.threads, nil
}

// Threads returns the cpu.TimesStat for each thread
func (s *Snapshot) ThreadsWithContext(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	return s.threads, nil
}

// Times returns CPU times of the process.
func (s *Snapshot) Times() (*cpu.TimesStat, error) {
	return s.lastCPUTimes, nil
}

// TimesWithContext returns CPU times of the process.
func (s *Snapshot) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	return s.lastCPUTimes, nil
}

// TODO: implement CPUAffinity

// MemoryInfo returns platform in-dependend memory information, such as RSS, VMS and Swap
func (s *Snapshot) MemoryInfo() (*MemoryInfoStat, error) {
	return s.memInfo, nil
}

// MemoryInfoWithContext returns platform in-dependend memory information, such as RSS, VMS and Swap
func (s *Snapshot) MemoryInfoWithContext() (*MemoryInfoStat, error) {
	return s.memInfo, nil
}

// MemoryInfoEx returns platform dependend memory information.
func (s *Snapshot) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return s.memInfoEx, nil
}

// MemoryInfoExWithContext returns platform dependend memory information.
func (s *Snapshot) MemoryInfoExWithContext() (*MemoryInfoExStat, error) {
	return s.memInfoEx, nil
}

// NOTE: explicitly not implementing Children command because it's 1:1

// OpenFiles returns a slice of OpenFilesStat opend by the process.
// OpenFilesStat includes a file path and file descriptor.
func (s *Snapshot) OpenFiles() ([]OpenFilesStat, error) {
	return s.OpenFilesWithContext(context.Background())
}

// OpenFiles returns a slice of OpenFilesStat opend by the process.
// OpenFilesStat includes a file path and file descriptor.
func (s *Snapshot) OpenFilesWithContext(ctx context.Context) ([]OpenFilesStat, error) {
	ret := make([]OpenFilesStat, len(s.openFiles))
	for i, o := range s.openFiles {
		ret[i] = *o
	}

	return ret, nil
}

// Connections returns a slice of net.ConnectionStat used by the process.
// This returns all kind of the connection. This measn TCP, UDP or UNIX.
func (s *Snapshot) Connections() ([]net.ConnectionStat, error) {
	return s.connections, nil
}

// ConnectionsWithContext returns a slice of net.ConnectionStat used by the process.
// This returns all kind of the connection. This measn TCP, UDP or UNIX.
func (s *Snapshot) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	return s.connections, nil
}

// NetIOCounters returns NetIOCounters of the process.
func (s *Snapshot) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	if pernic {
		return s.netIOCountersPerNic, nil
	}
	return s.netIOCounters, nil
}

// NetIOCountersWithContext returns NetIOCounters of the process.
func (s *Snapshot) NetIOCountersWithContext(ctx context.Context, pernic bool) ([]net.IOCountersStat, error) {
	if pernic {
		return s.netIOCountersPerNic, nil
	}
	return s.netIOCounters, nil
}

// TODO: implement IsRunning

// MemoryMaps returns memory maps from /proc/(pid)/smaps
func (s *Snapshot) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	if grouped {
		return s.memoryMapsGrouped, nil
	} else {
		return s.memoryMaps, nil
	}
}

// MemoryMapsWithContext returns memory maps from /proc/(pid)/smaps
func (s *Snapshot) MemoryMapsWithContext(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	if grouped {
		return s.memoryMapsGrouped, nil
	} else {
		return s.memoryMaps, nil
	}
}

// MemoryPercent returns how many percent of the total RAM this process uses
func (s *Snapshot) MemoryPercent() (float32, error) {
	return s.memPercent, nil
}

// MemoryPercentWithContext returns how many percent of the total RAM this process uses
func (s *Snapshot) MemoryPercentWithContext(ctx context.Context) (float32, error) {
	return s.memPercent, nil
}

// CPUPercent returns how many percent of the CPU time this process uses
func (s *Snapshot) CPUPercent() (float64, error) {
	return s.cpuPercent, nil
}

// CPUPercentWithContext returns how many percent of the CPU time this process uses
func (s *Snapshot) CPUPercentWithContext(ctx context.Context) (float64, error) {
	return s.cpuPercent, nil
}

// getFunctionsToCall returns a list of functions to call to make the snapshot
func getFunctionsToCall(fns ...Function) (functions []func(s *Snapshot, ctx context.Context) error) {
	functions = make([]func(s *Snapshot, ctx context.Context) error, len(internalFunctions))
	if len(fns) > 0 {
		for _, f := range fns {
			for k, fn := range functionToInternal[f] {
				functions[k] = fn
			}
		}
	} else {
		functions = internalFunctions
	}
	return
}

// Snapshot returns a snapshot of a process
func (p *Process) Snapshot(functions ...Function) (s *Snapshot, err error) {
	return p.SnapshotWithContext(context.Background(), functions...)
}

// SnapshotWithContext returns a snapshot of a process
func (p *Process) SnapshotWithContext(ctx context.Context, functions ...Function) (s *Snapshot, err error) {
	commonSnapshot, _ := NewCommonSnapshot(p.Pid)
	s = &Snapshot{
		CommonSnapshot: commonSnapshot,
	}
	ctx = context.Background()

	// get all of the functions that were specified
	fns := getFunctionsToCall(functions...)
	for _, fn := range fns {
		if fn != nil {
			if err = fn(s, ctx); err != nil {
				return
			}
		}
	}
	return
}

// Snapshots returns a list of snapshotted processes
func Snapshots(functions ...Function) (s []*Snapshot, err error) {
	if pids, err := Pids(); err == nil {
		s = make([]*Snapshot, len(pids))
		for _, pid := range pids {
			if p, skiperr := NewProcess(pid); skiperr == nil {
				if snap, skiperr := p.Snapshot(); skiperr == nil {
					s = append(s, snap)
				}
			}
		}
	}
	return
}
