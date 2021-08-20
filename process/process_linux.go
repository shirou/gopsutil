// +build linux

package process

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/cihub/seelog"

	"github.com/DataDog/gopsutil/cpu"
	"github.com/DataDog/gopsutil/host"
	"github.com/DataDog/gopsutil/internal/common"
	"github.com/DataDog/gopsutil/net"
)

var (
	CachedBootTime  = uint64(0)
	ErrorNoChildren = errors.New("process does not have children")
	PageSize        = uint64(os.Getpagesize())
)

const (
	PrioProcess               = 0   // linux/resource.h
	ClockTicks                = 100 // C.sysconf(C._SC_CLK_TCK)
	WorldReadable os.FileMode = 4
)

type MemoryMapSectionPermissions struct {
	Read    bool
	Write   bool
	Execute bool
	Private bool
	Shared  bool
}

func (msp *MemoryMapSectionPermissions) parse(perm string) error {
	if len(perm) != 4 {
		return fmt.Errorf("can't parse section permission '%s'", perm)
	}
	for _, v := range perm {
		switch v {
		case 'r':
			msp.Read = true
		case 'w':
			msp.Write = true
		case 'x':
			msp.Execute = true
		case 'p':
			msp.Private = true
		case 's':
			msp.Shared = true
		}
	}
	return nil
}

type MemoryMapSectionHeader struct {
	StartAddr  uintptr
	EndAddr    uintptr
	Permission MemoryMapSectionPermissions
	Offset     uint64
	Device     string
	Inode      uint64
}

type MemoryMapsStat struct {
	MemoryMapSectionHeader

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

func (m *MemoryMapsStat) updateAttribute(fields []string) error {
	if len(fields) < 2 {
		return fmt.Errorf("update attribute not enough fields")
	}
	var v *uint64
	switch fields[0] {
	case "Size:":
		v = &m.Size
	case "Rss:":
		v = &m.Rss
	case "Pss:":
		v = &m.Pss
	case "Shared_Clean:":
		v = &m.SharedClean
	case "Shared_Dirty:":
		v = &m.SharedDirty
	case "Private_Clean:":
		v = &m.PrivateClean
	case "Private_Dirty:":
		v = &m.PrivateDirty
	case "Referenced:":
		v = &m.Referenced
	case "Anonymous:":
		v = &m.Anonymous
	case "Swap:":
		v = &m.Swap
	default:
		// ignore other fields
		return nil
	}

	t, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return err
	}

	*v = t
	return nil
}

/*
   Parse a memory section from /proc/pid/maps or header section line of /proc/pid/smaps

 StartAddr        EndAddr    Permission Offset Device Inode              Path
ffffffffff600000-ffffffffff601000 --xp 00000000 00:00 0                  [vsyscall]
*/
func (ms *MemoryMapsStat) parseHeader(fields []string) error {
	if len(fields) < 5 {
		return fmt.Errorf("can't parse section, not enough fields '%v'", fields)
	}

	/* Address Range */
	i := strings.IndexRune(fields[0], '-')
	if i < 0 {
		return fmt.Errorf("can't parse address section, not enough fields '%v'", fields[0])
	}
	startAddr := fields[0][:i]
	endAddr := fields[0][i+1:]
	a, err := strconv.ParseUint(startAddr, 16, 64)
	if err != nil {
		return err
	}
	ms.StartAddr = uintptr(a)
	a, err = strconv.ParseUint(endAddr, 16, 64)
	if err != nil {
		return err
	}
	ms.EndAddr = uintptr(a)

	// Permission
	if err := ms.Permission.parse(fields[1]); err != nil {
		return err
	}

	// Offset
	a, err = strconv.ParseUint(fields[2], 16, 64)
	if err != nil {
		return err
	}
	ms.Offset = a

	// Device
	dev := fields[3]
	if len(dev) < 5 || !strings.Contains(dev, ":") {
		return fmt.Errorf("can't parse device section, '%s'", dev)
	}
	ms.Device = dev

	// Inode
	a, err = strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return err
	}
	ms.Inode = a

	if len(fields) >= 6 {
		ms.Path = strings.Join(fields[5:], " ")
	}

	return nil
}

func (ms *MemoryMapsStat) IsAnonymous() bool {
	return len(ms.Path) == 0
}

type currentUser struct {
	uid  uint32
	euid uint32
}

// String returns JSON value of the process.
func (m MemoryMapsStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// NewProcess creates a new Process instance, it only stores the pid and
// checks that the process exists. Other method on Process can be used
// to get more information about the process. An error will be returned
// if the process does not exist.
func NewProcess(pid int32) (*Process, error) {
	p := &Process{
		Pid: int32(pid),
	}
	file, err := os.Open(common.HostProc(strconv.Itoa(int(p.Pid))))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return p, err
}

// Ppid returns Parent Process ID of the process.
func (p *Process) Ppid() (int32, error) {
	ppid, _, _, _, _, err := p.fillFromStat()
	if err != nil {
		return -1, err
	}
	return ppid, nil
}

// Name returns name of the process.
func (p *Process) Name() (string, error) {
	if p.name == "" {
		if err := p.fillFromStatus(); err != nil {
			return "", err
		}
	}
	return p.name, nil
}

// Exe returns executable path of the process.
func (p *Process) Exe() (string, error) {
	return p.fillFromExe(getCurrentUser())
}

// Cmdline returns the command line arguments of the process as a string with
// each argument separated by 0x20 ascii character.
func (p *Process) Cmdline() (string, error) {
	return p.fillFromCmdline()
}

// CmdlineSlice returns the command line arguments of the process as a slice with each
// element being an argument.
func (p *Process) CmdlineSlice() ([]string, error) {
	return p.fillSliceFromCmdline()
}

// CreateTime returns created time of the process in seconds since the epoch, in UTC.
func (p *Process) CreateTime() (int64, error) {
	_, _, _, createTime, _, err := p.fillFromStat()
	if err != nil {
		return 0, err
	}
	return createTime, nil
}

// Cwd returns current working directory of the process.
func (p *Process) Cwd() (string, error) {
	return p.fillFromCwd(getCurrentUser())
}

// Parent returns parent Process of the process.
func (p *Process) Parent() (*Process, error) {
	err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	if p.parent == 0 {
		return nil, fmt.Errorf("wrong number of parents")
	}
	return NewProcess(p.parent)
}

// Status returns the process status.
// Return value could be one of these.
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
// The charactor is same within all supported platforms.
func (p *Process) Status() (string, error) {
	err := p.fillFromStatus()
	if err != nil {
		return "", err
	}
	return p.status, nil
}

// Uids returns user ids of the process as a slice of the int
func (p *Process) Uids() ([]int32, error) {
	err := p.fillFromStatus()
	if err != nil {
		return []int32{}, err
	}
	return p.uids, nil
}

// Gids returns group ids of the process as a slice of the int
func (p *Process) Gids() ([]int32, error) {
	err := p.fillFromStatus()
	if err != nil {
		return []int32{}, err
	}
	return p.gids, nil
}

// Terminal returns a terminal which is associated with the process.
func (p *Process) Terminal() (string, error) {
	terminal, err := p.fillTermFromStat()
	if err != nil {
		return "", err
	}
	return terminal, nil
}

// Nice returns a nice value (priority).
// Notice: gopsutil can not set nice value.
func (p *Process) Nice() (int32, error) {
	_, _, _, _, nice, err := p.fillFromStat()
	if err != nil {
		return 0, err
	}
	return nice, nil
}

// IOnice returns process I/O nice value (priority).
func (p *Process) IOnice() (int32, error) {
	return 0, common.ErrNotImplementedError
}

// Rlimit returns Resource Limits.
func (p *Process) Rlimit() ([]RlimitStat, error) {
	return nil, common.ErrNotImplementedError
}

// IOCounters returns IO Counters.
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return p.fillFromIO(getCurrentUser())
}

// NumCtxSwitches returns the number of the context switches of the process.
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	return p.numCtxSwitches, nil
}

// NumFDs returns the number of File Descriptors used by the process.
func (p *Process) NumFDs() (int32, error) {
	_, fds, err := p.fillFromfdList(getCurrentUser())
	return int32(len(fds)), err
}

// NumThreads returns the number of threads used by the process.
func (p *Process) NumThreads() (int32, error) {
	err := p.fillFromStatus()
	if err != nil {
		return 0, err
	}
	return p.numThreads, nil
}

// Threads returns a map of threads
//
// Notice: Not implemented yet. always returns empty map.
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, nil
}

// Times returns CPU times of the process.
func (p *Process) Times() (*cpu.TimesStat, error) {
	_, _, cpuTimes, _, _, err := p.fillFromStat()
	if err != nil {
		return nil, err
	}
	return cpuTimes, nil
}

// CPUAffinity returns CPU affinity of the process.
//
// Notice: Not implemented yet.
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, common.ErrNotImplementedError
}

// MemoryInfo returns platform in-dependend memory information, such as RSS, VMS and Swap
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	meminfo, _, err := p.readFromStatm()
	if err != nil {
		return nil, err
	}
	return meminfo, nil
}

// MemoryInfoEx returns platform dependend memory information.
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	_, memInfoEx, err := p.readFromStatm()
	if err != nil {
		return nil, err
	}
	return memInfoEx, nil
}

// Children returns a slice of Process of the process.
func (p *Process) Children() ([]*Process, error) {
	pids, err := common.CallPgrep(invoke, p.Pid)
	if err != nil {
		if pids == nil || len(pids) == 0 {
			return nil, ErrorNoChildren
		}
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

// OpenFiles returns a slice of OpenFilesStat opend by the process.
// OpenFilesStat includes a file path and file descriptor.
func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	_, ofs, err := p.fillFromfd(getCurrentUser())
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
	return net.ConnectionsPid("all", p.Pid)
}

// NetIOCounters returns NetIOCounters of the process.
func (p *Process) NetIOCounters(pernic bool) ([]net.IOCountersStat, error) {
	filename := common.HostProc(strconv.Itoa(int(p.Pid)), "net/dev")
	return net.IOCountersByFile(pernic, filename)
}

// IsRunning returns whether the process is running or not.
// Not implemented yet.
func (p *Process) IsRunning() (bool, error) {
	return true, common.ErrNotImplementedError
}

// MemoryMaps get memory maps from /proc/(pid)/smaps
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	pid := p.Pid
	var memMaps []MemoryMapsStat
	smapsPath := common.HostProc(strconv.Itoa(int(pid)), "smaps")

	smaps, err := os.Open(smapsPath)
	if err != nil {
		return nil, err
	}
	defer smaps.Close()

	scanner := bufio.NewScanner(bufio.NewReader(smaps))

	firstLine := true
	mms := MemoryMapsStat{}
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Fields(line)
		if !strings.HasSuffix(cols[0], ":") {
			if firstLine {
				firstLine = false
			} else {
				memMaps = append(memMaps, mms)
			}
			mms = MemoryMapsStat{}
			if err := mms.parseHeader(cols); err != nil {
				return nil, err
			}
		} else {
			if err := mms.updateAttribute(cols); err != nil {
				return nil, err
			}
		}
	}
	if firstLine {
		return nil, fmt.Errorf("the proc/pid/smaps parsing is empty. smaps path : %s", smapsPath)
	}
	memMaps = append(memMaps, mms)
	return &memMaps, nil
}

/**
** Internal functions
**/

// Get list of /proc/(pid)/fd files
func (p *Process) fillFromfdList(user *currentUser) (string, []string, error) {
	pid := p.Pid
	statPath := common.HostProc(strconv.Itoa(int(pid)), "fd")

	if err := ensurePathReadable(statPath, user); err != nil {
		return statPath, []string{}, err
	}

	d, err := os.Open(statPath)
	if err != nil {
		return statPath, []string{}, err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	return statPath, fnames, err
}

// Get num_fds from /proc/(pid)/fd
func (p *Process) fillFromfd(user *currentUser) (int32, []*OpenFilesStat, error) {
	statPath, fnames, err := p.fillFromfdList(user)
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
func (p *Process) fillFromCwd(user *currentUser) (string, error) {
	pid := p.Pid
	cwdPath := common.HostProc(strconv.Itoa(int(pid)), "cwd")
	if err := ensurePathReadable(cwdPath, user); err != nil {
		return "", err
	}
	cwd, err := os.Readlink(cwdPath)
	if err != nil {
		return "", err
	}
	return string(cwd), nil
}

// Get exe from /proc/(pid)/exe
func (p *Process) fillFromExe(user *currentUser) (string, error) {
	pid := p.Pid
	exePath := common.HostProc(strconv.Itoa(int(pid)), "exe")
	if err := ensurePathReadable(exePath, user); err != nil {
		return "", err
	}
	exe, err := os.Readlink(exePath)
	if err != nil {
		return "", err
	}
	return string(exe), nil
}

// Get cmdline from /proc/(pid)/cmdline
func (p *Process) fillFromCmdline() (string, error) {
	pid := p.Pid
	cmdPath := common.HostProc(strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return "", err
	}
	ret := strings.FieldsFunc(string(cmdline), func(r rune) bool {
		if r == '\u0000' {
			return true
		}
		return false
	})

	return strings.Join(ret, " "), nil
}

func (p *Process) fillSliceFromCmdline() ([]string, error) {
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
func (p *Process) fillFromIO(user *currentUser) (*IOCountersStat, error) {
	pid := p.Pid
	ioPath := common.HostProc(strconv.Itoa(int(pid)), "io")

	if err := ensurePathReadable(ioPath, user); err != nil {
		return nil, err
	}

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
func (p *Process) readFromStatm() (*MemoryInfoStat, *MemoryInfoExStat, error) {
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
func (p *Process) fillFromStatus() error {
	pid := p.Pid

	// make sure ctxSwitches/memInfo are not nil since we are about to fill them in.
	if p.numCtxSwitches == nil {
		p.numCtxSwitches = &NumCtxSwitchesStat{}
	}
	if p.memInfo == nil {
		p.memInfo = &MemoryInfoStat{}
	}

	statPath := common.HostProc(strconv.Itoa(int(pid)), "status")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		value := tabParts[1]
		switch strings.TrimRight(tabParts[0], ":") {
		case "Name":
			p.name = strings.Trim(value, " \t")
		case "State":
			p.status = value[0:1]
		case "PPid", "Ppid":
			pval, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return err
			}
			p.parent = int32(pval)
		case "Uid":
			p.uids = make([]int32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					return err
				}
				p.uids = append(p.uids, int32(v))
			}
		case "Gid":
			p.gids = make([]int32, 0, 4)
			for _, i := range strings.Split(value, "\t") {
				v, err := strconv.ParseInt(i, 10, 32)
				if err != nil {
					return err
				}
				p.gids = append(p.gids, int32(v))
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

		case "NSpid":
			values := strings.Split(value, "\t")
			// only report process namespaced PID
			v, err := strconv.ParseInt(values[len(values)-1], 10, 32)
			if err != nil {
				return err
			}
			p.NsPid = int32(v)
		}

	}
	return nil
}

func (p *Process) fillTermFromStat() (string, error) {
	pid := p.Pid
	statPath := common.HostProc(strconv.Itoa(int(pid)), "stat")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(contents))

	i := 1
	for !strings.HasSuffix(fields[i], ")") {
		i++
	}

	termmap, err := getTerminalMap()
	terminal := ""
	if err == nil {
		t, err := strconv.ParseUint(fields[i+5], 10, 64)
		if err != nil {
			return "", err
		}
		terminal = termmap[t]
	}
	return terminal, err
}

func (p *Process) fillFromStat() (int32, int32, *cpu.TimesStat, int64, int32, error) {
	pid := p.Pid
	statPath := common.HostProc(strconv.Itoa(int(pid)), "stat")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}
	fields := strings.Fields(string(contents))
	timestamp := time.Now().Unix()

	i := 1
	for !strings.HasSuffix(fields[i], ")") {
		i++
	}

	ppid, err := strconv.ParseInt(fields[i+2], 10, 32)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}
	pgrp, err := strconv.ParseInt(fields[i+3], 10, 32)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}
	utime, err := strconv.ParseFloat(fields[i+12], 64)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}

	stime, err := strconv.ParseFloat(fields[i+13], 64)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}

	cpuTimes := &cpu.TimesStat{
		CPU:       "cpu",
		User:      float64(utime / ClockTicks),
		System:    float64(stime / ClockTicks),
		Timestamp: timestamp,
	}

	if CachedBootTime == 0 {
		btime, _ := host.BootTime()
		CachedBootTime = btime
	}
	bootTime := CachedBootTime
	t, err := strconv.ParseUint(fields[i+20], 10, 64)
	if err != nil {
		return 0, 0, nil, 0, 0, err
	}
	ctime := (t / uint64(ClockTicks)) + uint64(bootTime)
	createTime := int64(ctime * 1000)

	//	p.Nice = mustParseInt32(fields[18])
	// use syscall instead of parse Stat file
	snice, _ := syscall.Getpriority(PrioProcess, int(pid))
	nice := int32(snice) // FIXME: is this true?

	return int32(ppid), int32(pgrp), cpuTimes, createTime, nice, nil
}

// Pids returns a slice of process ID list which are running now.
func Pids() ([]int32, error) {
	var ret []int32

	d, err := os.Open(common.HostProc())
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

func AllProcesses() (map[int32]*FilledProcess, error) {
	pids, err := Pids()
	if err != nil {
		return nil, fmt.Errorf("could not collect pids: %s", err)
	}

	user := getCurrentUser()
	// Collect stats on all processes limiting the number of required syscalls.
	// All errors are swallowed for now.
	procs := make(map[int32]*FilledProcess)
	for _, pid := range pids {
		p, err := NewProcess(pid)
		if err != nil {
			log.Debugf("Unable to create new process %d, it may have gone away: %s", pid, err)
			// Skip the rest of the processing because we have no real process.
			continue
		}
		cmdline, err := p.fillSliceFromCmdline()
		if err != nil {
			log.Debugf("Unable to read process command line for %d: %s", pid, err)
			cmdline = []string{}
		}
		if err := p.fillFromStatus(); err != nil {
			log.Debugf("Unable to fill from /proc/%d/status: %s", pid, err)
		}
		memInfo, memInfoEx, err := p.readFromStatm()
		if err != nil {
			log.Debugf("Unable to fill from /proc/%d/statm: %s", pid, err)
			memInfo = &MemoryInfoStat{}
			memInfoEx = &MemoryInfoExStat{}
		}
		ioStat, err := p.fillFromIO(user)
		if os.IsPermission(err) {
			log.Tracef("Unable to access /proc/%d/io, permission denied", pid)
			// Without root permissions we can't read for other processes.
			ioStat = &IOCountersStat{}
		} else if err != nil {
			log.Debugf("Unable to access /proc/%d/io: %s", pid, err)
			ioStat = &IOCountersStat{}
		}
		ppid, _, t1, createTime, nice, err := p.fillFromStat()
		if err != nil {
			log.Debugf("Unable to fill from /proc/%d/stat: %s", pid, err)
			t1 = &cpu.TimesStat{}
		}
		cwd, err := p.fillFromCwd(user)
		if os.IsPermission(err) {

			log.Tracef("Unable to access /proc/%d/cwd, permission denied", pid)
			cwd = ""
		} else if err != nil {
			log.Debugf("Unable to access /proc/%d/cwd: %s", pid, err)
			cwd = ""
		}
		exe, err := p.fillFromExe(user)
		if os.IsPermission(err) {
			// Without root permissions we can't read for other processes.
			log.Tracef("Unable to access /proc/%d/exe, permission denied", pid)
			exe = ""
		} else if err != nil {
			log.Debugf("Unable to access /proc/%d/exe: %s", pid, err)
			exe = ""
		}
		openFdCount := int32(-1)
		_, fds, err := p.fillFromfdList(user)
		if os.IsPermission(err) {
			// Without root permissions we can't read for other processes.
			log.Tracef("Unable to access /proc/%d/fd, permission denied", pid)
		} else if err != nil {
			log.Debugf("Unable to access /proc/%d/fd: %s", pid, err)
		} else {
			openFdCount = int32(len(fds))
		}

		procs[p.Pid] = &FilledProcess{
			Pid:     pid,
			Ppid:    ppid,
			NsPid:   p.NsPid,
			Cmdline: cmdline,
			// stat
			CpuTime:     *t1,
			Nice:        nice,
			CreateTime:  createTime,
			OpenFdCount: openFdCount,

			// status
			Name:        p.name,
			Status:      p.status,
			Uids:        p.uids,
			Gids:        p.gids,
			NumThreads:  p.numThreads,
			CtxSwitches: p.numCtxSwitches,
			// statm
			MemInfo:   memInfo,
			MemInfoEx: memInfoEx,
			// cwd
			Cwd: cwd,
			// exe
			Exe: exe,
			// IO
			IOStat: ioStat,
		}
	}
	return procs, nil
}

func getCurrentUser() *currentUser {
	return &currentUser{
		uid:  uint32(os.Getuid()),
		euid: uint32(os.Geteuid()),
	}
}

// ensurePathReadable ensures that the current user is able to read the path in question
// before opening it.  on some systems, attempting to open a file that the user does
// not have permission is problematic for customer security auditing
func ensurePathReadable(path string, user *currentUser) error {
	// User is (effectively or actually) root
	if user.euid == 0 {
		return nil
	}

	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	// File mode is world readable and not a symlink
	// If the file is a symlink, the owner check below will cover it
	if mode := info.Mode(); mode&os.ModeSymlink == 0 && mode.Perm()&WorldReadable != 0 {
		return nil
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		// If file is not owned by the user id or effective user id, return a permission error
		// Group permissions don't come in to play with procfs so we don't bother checking
		if stat.Uid != user.uid && stat.Uid != user.euid {
			return os.ErrPermission
		}
	}

	return nil
}
