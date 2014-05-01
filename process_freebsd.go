// +build freebsd

package gopsutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"
)

// MemoryInfoExStat is different between OSes
type MemoryInfoExStat struct {
}

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

	return k.KiPpid, nil
}
func (p *Process) Name() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return string(k.KiComm[:]), nil
}
func (p *Process) Exe() (string, error) {
	return "", errors.New("not implemented yet")
}
func (p *Process) Cmdline() (string, error) {
	return "", errors.New("not implemented yet")
}
func (p *Process) CreateTime() (int64, error) {
	return 0, errors.New("not implemented yet")
}
func (p *Process) Cwd() (string, error) {
	return "", errors.New("not implemented yet")
}
func (p *Process) Parent() (*Process, error) {
	return p, errors.New("not implemented yet")
}
func (p *Process) Status() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return string(k.KiStat[:]), nil
}
func (p *Process) Username() (string, error) {
	return "", errors.New("not implemented yet")
}
func (p *Process) Uids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	uids := make([]int32, 0, 3)

	uids = append(uids, int32(k.KiRuid), int32(k.KiUID), int32(k.KiSvuid))

	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]int32, 0, 3)
	gids = append(gids, int32(k.KiRgid), int32(k.KiNgroups[0]), int32(k.KiSvuid))

	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	ttyNr := uint64(k.KiTdev)

	termmap, err := getTerminalMap()
	if err != nil {
		return "", err
	}

	return termmap[ttyNr], nil
}
func (p *Process) Nice() (int32, error) {
	return 0, errors.New("not implemented yet")
}
func (p *Process) IOnice() (int32, error) {
	return 0, errors.New("not implemented yet")
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, errors.New("not implemented yet")
}
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return nil, errors.New("not implemented yet")
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return nil, errors.New("not implemented yet")
}
func (p *Process) NumFDs() (int32, error) {
	return 0, errors.New("not implemented yet")
}
func (p *Process) NumThreads() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.KiNumthreads, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, errors.New("not implemented yet")
}
func (p *Process) CPUTimes() (*CPUTimesStat, error) {
	return nil, errors.New("not implemented yet")
}
func (p *Process) CPUPercent() (int32, error) {
	return 0, errors.New("not implemented yet")
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, errors.New("not implemented yet")
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	ret := &MemoryInfoStat{
		RSS: uint64(k.KiRssize),
		VMS: uint64(k.KiSize),
	}

	return ret, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return nil, errors.New("not implemented yet")
}
func (p *Process) MemoryPercent() (float32, error) {
	return 0, errors.New("not implemented yet")
}

func (p *Process) Children() ([]*Process, error) {
	return nil, errors.New("not implemented yet")
}

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return nil, errors.New("not implemented yet")
}

func (p *Process) Connections() ([]NetConnectionStat, error) {
	return nil, errors.New("not implemented yet")
}

func (p *Process) IsRunning() (bool, error) {
	return true, errors.New("not implemented yet")
}
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	var ret []MemoryMapsStat
	return &ret, errors.New("not implemented yet")
}

func copyParams(k *KinfoProc, p *Process) error {

	return nil
}

func processes() ([]Process, error) {
	results := make([]Process, 0, 50)

	mib := []int32{CTL_KERN, KERN_PROC, KERN_PROC_PROC, 0}
	buf, length, err := callSyscall(mib)
	if err != nil {
		return results, err
	}

	// get kinfo_proc size
	k := KinfoProc{}
	procinfoLen := int(unsafe.Sizeof(k))
	count := int(length / uint64(procinfoLen))

	// parse buf to procs
	for i := 0; i < count; i++ {
		b := buf[i*procinfoLen : i*procinfoLen+procinfoLen]
		k, err := parseKinfoProc(b)
		if err != nil {
			continue
		}
		p, err := NewProcess(int32(k.KiPid))
		if err != nil {
			continue
		}
		copyParams(&k, p)

		results = append(results, *p)
	}

	return results, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.LittleEndian, &k)
	if err != nil {
		return k, err
	}

	return k, nil
}

func callSyscall(mib []int32) ([]byte, uint64, error) {
	miblen := uint64(len(mib))

	// get required buffer size
	length := uint64(0)
	_, _, err := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		0,
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		var b []byte
		return b, length, err
	}
	if length == 0 {
		var b []byte
		return b, length, err
	}
	// get proc info itself
	buf := make([]byte, length)
	_, _, err = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if err != 0 {
		return buf, length, err
	}

	return buf, length, nil
}

func (p *Process) getKProc() (*KinfoProc, error) {
	mib := []int32{CTL_KERN, KERN_PROC, KERN_PROC_PID, p.Pid}

	buf, length, err := callSyscall(mib)
	if err != nil {
		return nil, err
	}
	procK := KinfoProc{}
	if length != uint64(unsafe.Sizeof(procK)) {
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
