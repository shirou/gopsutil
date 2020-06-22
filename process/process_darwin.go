// +build darwin

package process

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/internal/common"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"golang.org/x/sys/unix"
)

// copied from sys/sysctl.h
const (
	CTLKern          = 1  // "high kernel": proc, limits
	KernProc         = 14 // struct: process entries
	KernProcPID      = 1  // by process id
	KernProcProc     = 8  // only return procs
	KernProcAll      = 0  // everything
	KernProcPathname = 12 // path to executable
)

const (
	ClockTicks = 100 // C.sysconf(C._SC_CLK_TCK)
)

type _Ctype_struct___0 struct {
	Pad uint64
}

// MemoryInfoExStat is different between OSes
type MemoryInfoExStat struct {
}

type MemoryMapsStat struct {
}

func pidsWithContext(ctx context.Context) ([]int32, error) {
	var ret []int32

	pids, err := callPsWithContext(ctx, "pid", 0, false)
	if err != nil {
		return ret, err
	}

	for _, pid := range pids {
		v, err := strconv.Atoi(pid[0])
		if err != nil {
			return ret, err
		}
		ret = append(ret, int32(v))
	}

	return ret, nil
}

func (p *Process) Ppid() (int32, error) {
	return p.PpidWithContext(context.Background())
}

func (p *Process) PpidWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(FieldPpid) {
		return -1, ErrorFieldNotRequested
	}

	v, ok := p.cache["Ppid"].(valueOrError)
	if !ok {
		r, err := callPsWithContext(ctx, "ppid", p.Pid, false)
		if err != nil {
			return 0, err
		}

		return ppidFromPS(r[0], 0)
	}

	return v.value.(int32), v.err
}

func ppidFromPS(psLine []string, index int) (int32, error) {
	v, err := strconv.Atoi(psLine[index])
	if err != nil {
		return 0, err
	}

	return int32(v), err
}
func (p *Process) Name() (string, error) {
	return p.NameWithContext(context.Background())
}

func (p *Process) NameWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(FieldName) {
		return "", ErrorFieldNotRequested
	}

	k, err := p.getKProc()
	if err != nil {
		return "", err
	}
	name := common.IntToString(k.Proc.P_comm[:])

	if len(name) >= 15 {
		cmdlineSlice, err := p.cmdlineSliceWithContext(ctx)
		if err != nil {
			return "", err
		}
		if len(cmdlineSlice) > 0 {
			extendedName := filepath.Base(cmdlineSlice[0])
			if strings.HasPrefix(extendedName, p.name) {
				name = extendedName
			} else {
				name = cmdlineSlice[0]
			}
		}
	}

	return name, nil
}
func (p *Process) Tgid() (int32, error) {
	return 0, common.ErrNotImplementedError
}
func (p *Process) Exe() (string, error) {
	return p.ExeWithContext(context.Background())
}

// Cmdline returns the command line arguments of the process as a string with
// each argument separated by 0x20 ascii character.
func (p *Process) Cmdline() (string, error) {
	return p.CmdlineWithContext(context.Background())
}

func (p *Process) CmdlineWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(FieldCmdline) {
		return "", ErrorFieldNotRequested
	}

	ret, err := p.cmdlineSliceWithContext(ctx)
	return strings.Join(ret, " "), err
}

// CmdlineSlice returns the command line arguments of the process as a slice with each
// element being an argument. Because of current deficiencies in the way that the command
// line arguments are found, single arguments that have spaces in the will actually be
// reported as two separate items. In order to do something better CGO would be needed
// to use the native darwin functions.
func (p *Process) CmdlineSlice() ([]string, error) {
	return p.CmdlineSliceWithContext(context.Background())
}

func (p *Process) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	if !p.isFieldRequested(FieldCmdlineSlice) {
		return nil, ErrorFieldNotRequested
	}

	return p.cmdlineSliceWithContext(ctx)
}

func (p *Process) cmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	v, ok := p.cache["CmdlineSlice"].(valueOrError)
	if !ok {
		r, err := callPsWithContext(ctx, "command", p.Pid, false)
		if err != nil {
			return nil, err
		}
		return r[0], err
	}

	return v.value.([]string), v.err
}
func (p *Process) createTimeWithContext(ctx context.Context) (int64, error) {
	v, ok := p.cache["createTime"].(valueOrError)
	if !ok {
		r, err := callPsWithContext(ctx, "etime", p.Pid, false)
		if err != nil {
			return 0, err
		}

		return createTimeFromPS(r[0], 0)
	}

	return v.value.(int64), v.err
}

func createTimeFromPS(psLine []string, index int) (int64, error) {
	elapsedSegments := strings.Split(strings.Replace(psLine[index], "-", ":", 1), ":")
	var elapsedDurations []time.Duration
	for i := len(elapsedSegments) - 1; i >= 0; i-- {
		p, err := strconv.ParseInt(elapsedSegments[i], 10, 0)
		if err != nil {
			return 0, err
		}
		elapsedDurations = append(elapsedDurations, time.Duration(p))
	}

	var elapsed = time.Duration(elapsedDurations[0]) * time.Second
	if len(elapsedDurations) > 1 {
		elapsed += time.Duration(elapsedDurations[1]) * time.Minute
	}
	if len(elapsedDurations) > 2 {
		elapsed += time.Duration(elapsedDurations[2]) * time.Hour
	}
	if len(elapsedDurations) > 3 {
		elapsed += time.Duration(elapsedDurations[3]) * time.Hour * 24
	}

	start := time.Now().Add(-elapsed)
	return start.Unix() * 1000, nil
}
func (p *Process) Cwd() (string, error) {
	return p.CwdWithContext(context.Background())
}

func (p *Process) CwdWithContext(ctx context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}
func (p *Process) Parent() (*Process, error) {
	return p.ParentWithContext(context.Background())
}

func (p *Process) ParentWithContext(ctx context.Context) (*Process, error) {
	rr, err := common.CallLsofWithContext(ctx, invoke, p.Pid, "-FR")
	if err != nil {
		return nil, err
	}
	for _, r := range rr {
		if strings.HasPrefix(r, "p") { // skip if process
			continue
		}
		l := string(r)
		v, err := strconv.Atoi(strings.Replace(l, "R", "", 1))
		if err != nil {
			return nil, err
		}

		fields := make([]Field, 0, len(p.requestedFields))
		for f := range p.requestedFields {
			fields = append(fields, f)
		}

		if p.requestedFields != nil {
			return NewProcessWithFields(int32(v), fields...)
		}

		return NewProcess(int32(v))
	}
	return nil, fmt.Errorf("could not find parent line")
}
func (p *Process) Status() (string, error) {
	return p.StatusWithContext(context.Background())
}

func (p *Process) StatusWithContext(ctx context.Context) (string, error) {
	if !p.isFieldRequested(FieldStatus) {
		return "", ErrorFieldNotRequested
	}

	v, ok := p.cache["Status"].(valueOrError)

	if !ok {
		r, err := callPsWithContext(ctx, "state", p.Pid, false)
		if err != nil {
			return "", err
		}

		return statusFromPS(r[0], 0)
	}

	return v.value.(string), v.err
}

func statusFromPS(psLine []string, index int) (string, error) {
	return psLine[index][0:1], nil
}

func (p *Process) Foreground() (bool, error) {
	return p.ForegroundWithContext(context.Background())
}

func (p *Process) ForegroundWithContext(ctx context.Context) (bool, error) {
	if !p.isFieldRequested(FieldForeground) {
		return false, ErrorFieldNotRequested
	}

	return p.foregroundWithContext(ctx)
}

func (p *Process) foregroundWithContext(ctx context.Context) (bool, error) {
	// see https://github.com/shirou/gopsutil/issues/596#issuecomment-432707831 for implementation details
	v, ok := p.cache["foreground"].(valueOrError)
	if !ok {
		r, err := callPsWithContext(ctx, "stat", p.Pid, false)
		if err != nil {
			return false, err
		}

		return foregroundFromPS(r[0], 0)
	}

	return v.value.(bool), v.err
}

func foregroundFromPS(psLine []string, index int) (bool, error) {
	return strings.IndexByte(psLine[index], '+') != -1, nil
}

func (p *Process) Uids() ([]int32, error) {
	return p.UidsWithContext(context.Background())
}

func (p *Process) UidsWithContext(ctx context.Context) ([]int32, error) {
	if !p.isFieldRequested(FieldUids) {
		return nil, ErrorFieldNotRequested
	}

	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	// See: http://unix.superglobalmegacorp.com/Net2/newsrc/sys/ucred.h.html
	userEffectiveUID := int32(k.Eproc.Ucred.UID)

	return []int32{userEffectiveUID}, nil
}
func (p *Process) Gids() ([]int32, error) {
	return p.GidsWithContext(context.Background())
}

func (p *Process) GidsWithContext(ctx context.Context) ([]int32, error) {
	if !p.isFieldRequested(FieldGids) {
		return nil, ErrorFieldNotRequested
	}

	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]int32, 0, 3)
	gids = append(gids, int32(k.Eproc.Pcred.P_rgid), int32(k.Eproc.Ucred.Ngroups), int32(k.Eproc.Pcred.P_svgid))

	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	return p.TerminalWithContext(context.Background())
}

func (p *Process) TerminalWithContext(ctx context.Context) (string, error) {
	return "", common.ErrNotImplementedError
	/*
		k, err := p.getKProc()
		if err != nil {
			return "", err
		}

		ttyNr := uint64(k.Eproc.Tdev)
		termmap, err := getTerminalMap()
		if err != nil {
			return "", err
		}

		return termmap[ttyNr], nil
	*/
}
func (p *Process) Nice() (int32, error) {
	return p.NiceWithContext(context.Background())
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(FieldNice) {
		return 0, ErrorFieldNotRequested
	}

	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}
	return int32(k.Proc.P_nice), nil
}
func (p *Process) IOnice() (int32, error) {
	return p.IOniceWithContext(context.Background())
}

func (p *Process) IOniceWithContext(ctx context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	return p.RlimitWithContext(context.Background())
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, common.ErrNotImplementedError
}
func (p *Process) RlimitUsage(gatherUsed bool) ([]RlimitStat, error) {
	return p.RlimitUsageWithContext(context.Background(), gatherUsed)
}

func (p *Process) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, common.ErrNotImplementedError
}
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return p.IOCountersWithContext(context.Background())
}

func (p *Process) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return p.NumCtxSwitchesWithContext(context.Background())
}

func (p *Process) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	return nil, common.ErrNotImplementedError
}
func (p *Process) NumFDs() (int32, error) {
	return p.NumFDsWithContext(context.Background())
}

func (p *Process) NumFDsWithContext(ctx context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}
func (p *Process) NumThreads() (int32, error) {
	return p.NumThreadsWithContext(context.Background())
}

func (p *Process) NumThreadsWithContext(ctx context.Context) (int32, error) {
	if !p.isFieldRequested(FieldNumThreads) {
		return 0, ErrorFieldNotRequested
	}

	cacheKey := "NumThreads"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		r, err := callPsWithContext(ctx, "utime,stime", p.Pid, true)
		v = valueOrError{
			value: int32(len(r)),
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(int32), v.err
}
func (p *Process) Threads() (map[int32]*cpu.TimesStat, error) {
	return p.ThreadsWithContext(context.Background())
}

func (p *Process) ThreadsWithContext(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	ret := make(map[int32]*cpu.TimesStat)
	return ret, common.ErrNotImplementedError
}

func convertCPUTimes(s string) (ret float64, err error) {
	var t int
	var _tmp string
	if strings.Contains(s, ":") {
		_t := strings.Split(s, ":")
		switch len(_t) {
		case 3:
			hour, err := strconv.Atoi(_t[0])
			if err != nil {
				return ret, err
			}
			t += hour * 60 * 60 * ClockTicks

			mins, err := strconv.Atoi(_t[1])
			if err != nil {
				return ret, err
			}
			t += mins * 60 * ClockTicks
			_tmp = _t[2]
		case 2:
			mins, err := strconv.Atoi(_t[0])
			if err != nil {
				return ret, err
			}
			t += mins * 60 * ClockTicks
			_tmp = _t[1]
		case 1, 0:
			_tmp = s
		default:
			return ret, fmt.Errorf("wrong cpu time string")
		}
	} else {
		_tmp = s
	}

	_t := strings.Split(_tmp, ".")
	if err != nil {
		return ret, err
	}
	h, err := strconv.Atoi(_t[0])
	t += h * ClockTicks
	h, err = strconv.Atoi(_t[1])
	t += h
	return float64(t) / ClockTicks, nil
}
func (p *Process) Times() (*cpu.TimesStat, error) {
	return p.TimesWithContext(context.Background())
}

func (p *Process) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	if p.isFieldRequested(FieldTimes) {
		return nil, ErrorFieldNotRequested
	}

	v, ok := p.cache["Times"].(valueOrError)

	if !ok {
		r, err := callPsWithContext(ctx, "utime,stime", p.Pid, false)

		if err != nil {
			return nil, err
		}

		return timesFromPS(r[0], 0)
	}

	return v.value.(*cpu.TimesStat), v.err
}

func timesFromPS(psLine []string, index int) (*cpu.TimesStat, error) {
	utime, err := convertCPUTimes(psLine[index])
	if err != nil {
		return nil, err
	}
	stime, err := convertCPUTimes(psLine[index+1])
	if err != nil {
		return nil, err
	}

	ret := &cpu.TimesStat{
		CPU:    "cpu",
		User:   utime,
		System: stime,
	}
	return ret, nil
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return p.CPUAffinityWithContext(context.Background())
}

func (p *Process) CPUAffinityWithContext(ctx context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	return p.MemoryInfoWithContext(context.Background())
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {
	if !p.isFieldRequested(FieldMemoryInfo) {
		return nil, ErrorFieldNotRequested
	}

	v, ok := p.cache["MemoryInfo"].(valueOrError)

	if !ok {
		r, err := callPsWithContext(ctx, "rss,vsize,pagein", p.Pid, false)
		if err != nil {
			return nil, err
		}

		return memoryInfoFromPS(r[0], 0)
	}

	return v.value.(*MemoryInfoStat), v.err
}

func memoryInfoFromPS(psLine []string, index int) (*MemoryInfoStat, error) {
	rss, err := strconv.Atoi(psLine[index])
	if err != nil {
		return nil, err
	}
	vms, err := strconv.Atoi(psLine[index+1])
	if err != nil {
		return nil, err
	}
	pagein, err := strconv.Atoi(psLine[index+2])
	if err != nil {
		return nil, err
	}

	ret := &MemoryInfoStat{
		RSS:  uint64(rss) * 1024,
		VMS:  uint64(vms) * 1024,
		Swap: uint64(pagein),
	}

	return ret, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return p.MemoryInfoExWithContext(context.Background())
}

func (p *Process) MemoryInfoExWithContext(ctx context.Context) (*MemoryInfoExStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) PageFaults() (*PageFaultsStat, error) {
	return p.PageFaultsWithContext(context.Background())
}

func (p *Process) PageFaultsWithContext(ctx context.Context) (*PageFaultsStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) Children() ([]*Process, error) {
	return p.ChildrenWithContext(context.Background())
}

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
	pids, err := common.CallPgrepWithContext(ctx, invoke, p.Pid)
	if err != nil {
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

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return p.OpenFilesWithContext(context.Background())
}

func (p *Process) OpenFilesWithContext(ctx context.Context) ([]OpenFilesStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) Connections() ([]net.ConnectionStat, error) {
	return p.ConnectionsWithContext(context.Background())
}

func (p *Process) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	if !p.isFieldRequested(FieldConnections) {
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

func (p *Process) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	return p.NetIOCountersWithContext(context.Background(), pernic)
}

func (p *Process) NetIOCountersWithContext(ctx context.Context, pernic bool) ([]net.IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	return p.MemoryMapsWithContext(context.Background(), grouped)
}

func (p *Process) MemoryMapsWithContext(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	var ret []MemoryMapsStat
	return &ret, common.ErrNotImplementedError
}

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
		if f == FieldMemoryPercent {
			tmp, err := mem.VirtualMemory()
			if err == nil {
				machineMemory = tmp.Total
			}
		}
	}

	for _, pid := range pids {
		p, err := newProcessWithFields(pid, map[string]interface{}{"VirtualMemory": machineMemory}, fields...)
		if err != nil {
			continue
		}
		out = append(out, p)
	}

	return out, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	br := bytes.NewReader(buf)

	err := common.Read(br, binary.LittleEndian, &k)
	if err != nil {
		return k, err
	}

	return k, nil
}

// Returns a proc as defined here:
// http://unix.superglobalmegacorp.com/Net2/newsrc/sys/kinfo_proc.h.html
func (p *Process) getKProc() (*KinfoProc, error) {
	return p.getKProcWithContext(context.Background())
}

func (p *Process) getKProcWithContext(ctx context.Context) (*KinfoProc, error) {
	cacheKey := "getKProc"
	v, ok := p.cache[cacheKey].(valueOrError)

	if !ok {
		tmp, err := p.getKProcWithContextNoCache(ctx)
		v = valueOrError{
			value: tmp,
			err:   err,
		}
	}

	if p.cache != nil {
		p.cache[cacheKey] = v
	}

	return v.value.(*KinfoProc), v.err
}

func (p *Process) getKProcWithContextNoCache(ctx context.Context) (*KinfoProc, error) {
	mib := []int32{CTLKern, KernProc, KernProcPID, p.Pid}
	procK := KinfoProc{}
	length := uint64(unsafe.Sizeof(procK))
	buf := make([]byte, length)
	_, _, syserr := unix.Syscall6(
		unix.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if syserr != 0 {
		return nil, syserr
	}
	k, err := parseKinfoProc(buf)
	if err != nil {
		return nil, err
	}

	return &k, nil
}

// call ps command.
// Return value deletes Header line(you must not input wrong arg).
// And splited by Space. Caller have responsibility to manage.
// If passed arg pid is 0, get information from all process.
func callPsWithContext(ctx context.Context, arg string, pid int32, threadOption bool) ([][]string, error) {
	bin, err := exec.LookPath("ps")
	if err != nil {
		return [][]string{}, err
	}

	var cmd []string
	if pid == 0 { // will get from all processes.
		cmd = []string{"-ax", "-o", arg}
	} else if threadOption {
		cmd = []string{"-x", "-o", arg, "-M", "-p", strconv.Itoa(int(pid))}
	} else {
		cmd = []string{"-x", "-o", arg, "-p", strconv.Itoa(int(pid))}
	}
	out, err := invoke.CommandWithContext(ctx, bin, cmd...)
	if err != nil {
		return [][]string{}, err
	}
	lines := strings.Split(string(out), "\n")

	var ret [][]string
	for _, l := range lines[1:] {
		var lr []string
		for _, r := range strings.Split(l, " ") {
			if r == "" {
				continue
			}
			lr = append(lr, strings.TrimSpace(r))
		}
		if len(lr) != 0 {
			ret = append(ret, lr)
		}
	}

	return ret, nil
}

func (p *Process) prefetchFields(fields []Field) error {
	ctx := context.Background()

	fieldsNeedGeneric := make([]Field, 0, len(fields))
	psmultiFields := []string{"etime"}
	doCmdline := false

	for _, f := range fields {
		switch f {
		case FieldPpid:
			psmultiFields = append(psmultiFields, "ppid")
		case FieldCmdline, FieldCmdlineSlice:
			doCmdline = true
		case FieldCreateTime:
			// already done
		case FieldStatus:
			psmultiFields = append(psmultiFields, "state")
		case FieldForeground, FieldBackground:
			psmultiFields = append(psmultiFields, "stat")
		case FieldTimes:
			psmultiFields = append(psmultiFields, "utime,stime")
		case FieldMemoryInfo:
			psmultiFields = append(psmultiFields, "rss,vsize,pagein")
		default:
			fieldsNeedGeneric = append(fieldsNeedGeneric, f)
		}
	}

	if doCmdline {
		psmultiFields = append(psmultiFields, "command")
	}

	arg := strings.Join(psmultiFields, ",")
	r, err := callPsWithContext(ctx, arg, p.Pid, false)
	if err != nil {
		return err
	}

	for i, v := range strings.Split(arg, ",") {
		switch v {
		case "ppid":
			tmp, err := ppidFromPS(r[0], i)
			p.cache["Ppid"] = valueOrError{
				value: tmp,
				err:   err,
			}
		case "etime":
			tmp, err := createTimeFromPS(r[0], i)
			p.cache["createTime"] = valueOrError{
				value: tmp,
				err:   err,
			}
		case "state":
			tmp, err := statusFromPS(r[0], i)
			p.cache["Status"] = valueOrError{
				value: tmp,
				err:   err,
			}
		case "stat":
			tmp, err := foregroundFromPS(r[0], i)
			p.cache["foreground"] = valueOrError{
				value: tmp,
				err:   err,
			}
		case "utime":
			tmp, err := timesFromPS(r[0], i)
			p.cache["Times"] = valueOrError{
				value: tmp,
				err:   err,
			}
		case "rss":
			tmp, err := memoryInfoFromPS(r[0], i)
			p.cache["MemoryInfo"] = valueOrError{
				value: tmp,
				err:   err,
			}
		}
	}

	p.genericPrefetchFields(ctx, fieldsNeedGeneric)

	return nil
}
