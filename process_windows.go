// +build windows

package gopsutil

import (
	"errors"
	"github.com/shirou/w32"
	"syscall"
	"unsafe"
)

const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
)

type SYSTEM_PROCESS_INFORMATION struct {
	NextEntryOffset   uint64
	NumberOfThreads   uint64
	Reserved1         [48]byte
	Reserved2         [3]byte
	UniqueProcessId   uintptr
	Reserved3         uintptr
	HandleCount       uint64
	Reserved4         [4]byte
	Reserved5         [11]byte
	PeakPagefileUsage uint64
	PrivatePageCount  uint64
	Reserved6         [6]uint64
}

// Memory_info_ex is different between OSes
type Memory_info_exStat struct {
}

type Memory_mapsStat struct {
}

func Pids() ([]int32, error) {

	ret := make([]int32, 0)

	procs, err := processes()
	if err != nil {
		return ret, nil
	}
	for _, proc := range procs {
		ret = append(ret, proc.Pid)
	}

	return ret, nil
}

func (p *Process) Ppid() (int32, error) {
	ret, _, _, err := p.getFromSnapProcess(p.Pid)
	if err != nil {
		return 0, err
	}
	return ret, nil
}
func (p *Process) Name() (string, error) {
	name := ""
	return name, errors.New("Not implemented yet")
}
func (p *Process) Exe() (string, error) {
	_, _, ret, err := p.getFromSnapProcess(p.Pid)
	if err != nil {
		return "", err
	}
	return ret, nil
}
func (p *Process) Cmdline() (string, error) {
	return "", errors.New("Not implemented yet")
}
func (p *Process) Cwd() (string, error) {
	return "", errors.New("Not implemented yet")
}
func (p *Process) Parent() (*Process, error) {
	return p, errors.New("Not implemented yet")
}
func (p *Process) Status() (string, error) {
	return "", errors.New("Not implemented yet")
}
func (p *Process) Username() (string, error) {
	return "", errors.New("Not implemented yet")
}
func (p *Process) Uids() ([]int32, error) {
	uids := make([]int32, 0)
	return uids, errors.New("Not implemented yet")
}
func (p *Process) Gids() ([]int32, error) {
	gids := make([]int32, 0)
	return gids, errors.New("Not implemented yet")
}
func (p *Process) Terminal() (string, error) {
	return "", errors.New("Not implemented yet")
}
func (p *Process) Nice() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Ionice() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	rlimit := make([]RlimitStat, 0)
	return rlimit, errors.New("Not implemented yet")
}
func (p *Process) Io_counters() (*Io_countersStat, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Num_ctx_switches() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Num_fds() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Num_Threads() (int32, error) {
	_, ret, _, err := p.getFromSnapProcess(p.Pid)
	if err != nil {
		return 0, err
	}
	return ret, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, errors.New("Not implemented yet")
}
func (p *Process) Cpu_times() (*CPU_TimesStat, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Cpu_percent() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Cpu_affinity() ([]int32, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Memory_info() (*Memory_infoStat, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Memory_info_ex() (*Memory_info_exStat, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Memory_percent() (float32, error) {
	return 0, errors.New("Not implemented yet")
}

func (p *Process) Children() ([]*Process, error) {
	return nil, errors.New("Not implemented yet")
}

func (p *Process) Open_files() ([]Open_filesStat, error) {
	return nil, errors.New("Not implemented yet")
}

func (p *Process) Connections() ([]Net_connectionStat, error) {
	return nil, errors.New("Not implemented yet")
}

func (p *Process) Is_running() (bool, error) {
	return true, errors.New("Not implemented yet")
}

func (p *Process) Memory_Maps() (*[]Memory_mapsStat, error) {
	return nil, errors.New("Not implemented yet")
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	return p, nil
}

func (p *Process) Send_signal(sig syscall.Signal) error {
	return errors.New("Not implemented yet")
}

func (p *Process) Suspend() error {
	return errors.New("Not implemented yet")
}
func (p *Process) Resume() error {
	return errors.New("Not implemented yet")
}
func (p *Process) Terminate() error {
	return errors.New("Not implemented yet")
}
func (p *Process) Kill() error {
	return errors.New("Not implemented yet")
}

func (p *Process) getFromSnapProcess(pid int32) (int32, int32, string, error) {
	snap := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPPROCESS, uint32(pid))
	if snap == 0 {
		return 0, 0, "", syscall.GetLastError()
	}
	defer w32.CloseHandle(snap)
	var pe32 w32.PROCESSENTRY32
	pe32.DwSize = uint32(unsafe.Sizeof(pe32))
	if w32.Process32First(snap, &pe32) == false {
		return 0, 0, "", syscall.GetLastError()
	}

	if pe32.Th32ProcessID == uint32(pid) {
		szexe := syscall.UTF16ToString(pe32.SzExeFile[:])
		return int32(pe32.Th32ParentProcessID), int32(pe32.CntThreads), szexe, nil
	}

	for w32.Process32Next(snap, &pe32) {
		if pe32.Th32ProcessID == uint32(pid) {
			szexe := syscall.UTF16ToString(pe32.SzExeFile[:])
			return int32(pe32.Th32ParentProcessID), int32(pe32.CntThreads), szexe, nil
		}
	}
	return 0, 0, "", errors.New("Cloud not find pid:" + string(pid))
}

// Get processes
func processes() ([]*Process, error) {
	ps := make([]uint32, 255)
	var read uint32 = 0
	if w32.EnumProcesses(ps, uint32(len(ps)), &read) == false {
		return nil, syscall.GetLastError()
	}

	results := make([]*Process, 0)
	dward_size := uint32(4)
	for _, pid := range ps[:read/dward_size] {
		if pid == 0 {
			continue
		}
		p, err := NewProcess(int32(pid))
		if err != nil {
			break
		}
		results = append(results, p)
	}

	return results, nil
}

func get_proc_info(pid int32) (*SYSTEM_PROCESS_INFORMATION, error) {
	initialBufferSize := uint64(0x4000)
	bufferSize := initialBufferSize
	buffer := make([]byte, bufferSize)

	var sys_proc_info SYSTEM_PROCESS_INFORMATION
	ret, _, _ := procNtQuerySystemInformation.Call(
		uintptr(unsafe.Pointer(&sys_proc_info)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&bufferSize)),
		uintptr(unsafe.Pointer(&bufferSize)))
	if ret != 0 {
		return nil, syscall.GetLastError()
	}

	return &sys_proc_info, nil
}
