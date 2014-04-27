// +build windows

package gopsutil

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
	procNtQuerySystemInformation = modNt.NewProc("NtQuerySystemInformation")
)

const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
)

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [MAX_PATH]uint16
}

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
	return 0, nil
}
func (p *Process) Name() (string, error) {
	name := ""
	return name, nil
}
func (p *Process) Exe() (string, error) {
	return "", nil
}
func (p *Process) Cmdline() (string, error) {
	return "", nil
}
func (p *Process) Cwd() (string, error) {
	return "", nil
}
func (p *Process) Parent() (*Process, error) {
	return p, nil
}
func (p *Process) Status() (string, error) {
	return "", nil
}
func (p *Process) Username() (string, error) {
	return "", nil
}
func (p *Process) Uids() ([]int32, error) {
	uids := make([]int32, 0)
	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	gids := make([]int32, 0)
	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	return "", nil
}
func (p *Process) Nice() (int32, error) {
	return 0, nil
}
func (p *Process) Ionice() (int32, error) {
	return 0, nil
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	rlimit := make([]RlimitStat, 0)
	return rlimit, nil
}
func (p *Process) Io_counters() (*Io_countersStat, error) {
	return nil, nil
}
func (p *Process) Num_ctx_switches() (int32, error) {
	return 0, nil
}
func (p *Process) Num_fds() (int32, error) {
	return 0, nil
}
func (p *Process) Num_Threads() (int32, error) {
	return 0, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, nil
}
func (p *Process) Cpu_times() (*CPU_TimesStat, error) {
	return nil, nil
}
func (p *Process) Cpu_percent() (int32, error) {
	return 0, nil
}
func (p *Process) Cpu_affinity() ([]int32, error) {
	return nil, nil
}
func (p *Process) Memory_info() (*Memory_infoStat, error) {
	return nil, nil
}
func (p *Process) Memory_info_ex() (*Memory_info_exStat, error) {
	return nil, nil
}
func (p *Process) Memory_percent() (float32, error) {
	return 0, nil
}

func (p *Process) Children() ([]*Process, error) {
	return nil, nil
}

func (p *Process) Open_files() ([]Open_filesStat, error) {
	return nil, nil
}

func (p *Process) Connections() ([]Net_connectionStat, error) {
	return nil, nil
}

func (p *Process) Is_running() (bool, error) {
	return true, nil
}

func (p *Process) Memory_Maps() (*[]Memory_mapsStat, error) {
	return nil, nil
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	return p, nil
}

func (p *Process) Send_signal(sig syscall.Signal) error {
	return nil
}

func (p *Process) Suspend() error {
	return nil
}
func (p *Process) Resume() error {
	return nil
}
func (p *Process) Terminate() error {
	return nil
}
func (p *Process) Kill() error {
	return nil
}
func copy_params(pe32 *PROCESSENTRY32, p *Process) error {
	// p.Ppid = int32(pe32.Th32ParentProcessID)

	return nil
}

// Get processes
// This is borrowed from go-ps
func processes() ([]*Process, error) {
	handle, _, _ := procCreateToolhelp32Snapshot.Call(
		0x00000002,
		0)
	if handle < 0 {
		return nil, syscall.GetLastError()
	}
	defer procCloseHandle.Call(handle)

	var entry PROCESSENTRY32
	entry.DwSize = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Error retrieving process info.")
	}

	results := make([]*Process, 0)
	for {
		pid := int32(entry.Th32ProcessID)
		p, err := NewProcess(pid)
		if err != nil {
			break
		}
		results = append(results, p)

		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
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
