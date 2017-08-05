// +build freebsd

package process

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"syscall"

	cpu "github.com/DataDog/gopsutil/cpu"
	"github.com/DataDog/gopsutil/internal/common"
	net "github.com/DataDog/gopsutil/net"
)

type MemoryMapsStat struct {
}

func Pids() ([]int32, error) {
	var ret []int32
	procs, err := processes()
	if err != nil {
		return ret, nil
	}

	for _, p := range procs {
		ret = append(ret, p.Pid)
	}

	return ret, nil
}

func (p *Process) Ppid() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.Ppid, nil
}
func (p *Process) Name() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return common.IntToString(k.Comm[:]), nil
}
func (p *Process) Exe() (string, error) {
	return "", common.ErrNotImplementedError
}

func (p *Process) Cmdline() (string, error) {
	mib := []int32{CTLKern, KernProc, KernProcArgs, p.Pid}
	buf, _, err := common.CallSyscall(mib)
	if err != nil {
		return "", err
	}
	ret := strings.FieldsFunc(string(buf), func(r rune) bool {
		if r == '\u0000' {
			return true
		}
		return false
	})

	return strings.Join(ret, " "), nil
}

func (p *Process) CmdlineSlice() ([]string, error) {
	mib := []int32{CTLKern, KernProc, KernProcArgs, p.Pid}
	buf, _, err := common.CallSyscall(mib)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, nil
	}
	if buf[len(buf)-1] == 0 {
		buf = buf[:len(buf)-1]
	}
	parts := bytes.Split(buf, []byte{0})
	var strParts []string
	for _, p := range parts {
		strParts = append(strParts, string(p))
	}

	return strParts, nil
}
func (p *Process) CreateTime() (int64, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}
	return k.Start.Sec * 1000, nil
}
func (p *Process) Cwd() (string, error) {
	return "", common.ErrNotImplementedError
}
func (p *Process) Parent() (*Process, error) {
	return p, common.ErrNotImplementedError
}
func (p *Process) Status() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}
	return p.formatStatus(k.Stat), nil
}
func (p *Process) formatStatus(stat int8) string {
	var s string
	switch stat {
	case SIDL:
		s = "I"
	case SRUN:
		s = "R"
	case SSLEEP:
		s = "S"
	case SSTOP:
		s = "T"
	case SZOMB:
		s = "Z"
	case SWAIT:
		s = "W"
	case SLOCK:
		s = "L"
	}
	return s
}
func (p *Process) Uids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	uids := make([]int32, 0, 3)

	uids = append(uids, int32(k.Ruid), int32(k.Uid), int32(k.Svuid))

	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]int32, 0, 3)
	gids = append(gids, int32(k.Rgid), int32(k.Ngroups), int32(k.Svgid))

	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	ttyNr := uint64(k.Tdev)

	termmap, err := getTerminalMap()
	if err != nil {
		return "", err
	}

	return termmap[ttyNr], nil
}
func (p *Process) Nice() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}
	return int32(k.Nice), nil
}
func (p *Process) IOnice() (int32, error) {
	return 0, common.ErrNotImplementedError
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, common.ErrNotImplementedError
}
func (p *Process) IOCounters() (*IOCountersStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	return &IOCountersStat{
		ReadCount:  uint64(k.Rusage.Inblock),
		WriteCount: uint64(k.Rusage.Oublock),
	}, nil
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return nil, common.ErrNotImplementedError
}
func (p *Process) NumFDs() (int32, error) {
	return 0, common.ErrNotImplementedError
}
func (p *Process) NumThreads() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.Numthreads, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, common.ErrNotImplementedError
}
func (p *Process) Times() (*cpu.TimesStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	return &cpu.TimesStat{
		CPU:    "cpu",
		User:   float64(k.Rusage.Utime.Sec) + float64(k.Rusage.Utime.Usec)/1000000,
		System: float64(k.Rusage.Stime.Sec) + float64(k.Rusage.Stime.Usec)/1000000,
	}, nil
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, common.ErrNotImplementedError
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}
	return p.formatMemoryInfo(k)
}
func (p *Process) formatMemoryInfo(k *KinfoProc) (*MemoryInfoStat, error) {
	v, err := syscall.Sysctl("vm.stats.vm.v_page_size")
	if err != nil {
		return nil, err
	}
	pageSize := common.LittleEndian.Uint16([]byte(v))

	return &MemoryInfoStat{
		RSS: uint64(k.Rssize) * uint64(pageSize),
		VMS: uint64(k.Size),
	}, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) Children() ([]*Process, error) {
	pids, err := common.CallPgrep(invoke, p.Pid)
	if err != nil {
		return nil, err
	}
	ret := make([]*Process, 0, len(pids))
	for _, pid := range pids {
		np, err := NewProcess(pid)
		if err != nil {
			return nil, err
		}
		ret = append(ret, np)
	}
	return ret, nil
}

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) Connections() ([]net.ConnectionStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) IsRunning() (bool, error) {
	return true, common.ErrNotImplementedError
}
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	var ret []MemoryMapsStat
	return &ret, common.ErrNotImplementedError
}

func processes() ([]Process, error) {
	results := make([]Process, 0, 50)

	mib := []int32{CTLKern, KernProc, KernProcProc, 0}
	buf, length, err := common.CallSyscall(mib)
	if err != nil {
		return results, err
	}

	// get kinfo_proc size
	count := int(length / uint64(sizeOfKinfoProc))

	// parse buf to procs
	for i := 0; i < count; i++ {
		b := buf[i*sizeOfKinfoProc : (i+1)*sizeOfKinfoProc]
		k, err := parseKinfoProc(b)
		if err != nil {
			continue
		}
		p, err := NewProcess(int32(k.Pid))
		if err != nil {
			continue
		}

		results = append(results, *p)
	}

	return results, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	br := bytes.NewReader(buf)
	err := common.Read(br, binary.LittleEndian, &k)
	return k, err
}

func (p *Process) getKProc() (*KinfoProc, error) {
	mib := []int32{CTLKern, KernProc, KernProcPID, p.Pid}

	buf, length, err := common.CallSyscall(mib)
	if err != nil {
		return nil, err
	}
	if length != sizeOfKinfoProc {
		return nil, err
	}

	k, err := parseKinfoProc(buf)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	return p, nil
}

func AllProcesses() (map[int32]*FilledProcess, error) {
	pids, err := Pids()
	if err != nil {
		return nil, fmt.Errorf("could not collect pids: %s", err)
	}

	// Collect stats on all processes limiting the number of required syscalls.
	// All errors are swallowed for now.
	procs := make(map[int32]*FilledProcess)
	for _, pid := range pids {
		p, err := NewProcess(pid)
		if err != nil {
			continue
		}

		cmdline, err := p.CmdlineSlice()
		if err != nil {
			continue
		}
		k, err := p.getKProc()
		if err != nil {
			continue
		}
		memInfo, err := p.formatMemoryInfo(k)
		if err != nil {
			continue
		}

		procs[p.Pid] = &FilledProcess{
			Pid:     pid,
			Ppid:    k.Ppid,
			Cmdline: cmdline,
			CpuTime: cpu.TimesStat{
				CPU:    "cpu",
				User:   float64(k.Rusage.Utime.Sec) + float64(k.Rusage.Utime.Usec)/1000000,
				System: float64(k.Rusage.Stime.Sec) + float64(k.Rusage.Stime.Usec)/1000000,
			},
			Nice:        int32(k.Nice),
			Status:      p.formatStatus(k.Stat),
			CreateTime:  k.Start.Sec * 1000,
			OpenFdCount: 0, // FIXME: Needs implementation
			Name:        common.IntToString(k.Comm[:]),
			Uids:        []int32{int32(k.Ruid), int32(k.Uid), int32(k.Svuid)},
			Gids:        []int32{int32(k.Rgid), int32(k.Ngroups), int32(k.Svgid)},
			NumThreads:  k.Numthreads,
			MemInfo:     memInfo,
			Cwd:         "",
			IOStat: &IOCountersStat{
				ReadCount:  uint64(k.Rusage.Inblock),
				WriteCount: uint64(k.Rusage.Oublock),
			},
		}
	}
	return procs, nil
}
