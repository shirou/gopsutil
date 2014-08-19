// +build linux

package gopsutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	PrioProcess = 0 // linux/resource.h
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
	SharedClean  uint64 `json:"shared_clean"`
	SharedDirty  uint64 `json:"shared_dirty"`
	PrivateClean uint64 `json:"private_clean"`
	PrivateDirty uint64 `json:"private_dirty"`
	Referenced   uint64 `json:"referenced"`
	Anonymous    uint64 `json:"anonymous"`
	Swap         uint64 `json:"swap"`
}

func (m MemoryMapsStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

// Create new Process instance
// This only stores Pid
func NewProcess(pid int32) (*Process, error) {
	p := &Process{
		Pid: int32(pid),
	}
	return p, nil
}

func (p *Process) Ppid() (int32, error) {
	_, ppid, _, _, _, err := p.fillFromStat()
	if err != nil {
		return -1, err
	}
	return ppid, nil
}
func (p *Process) Name() (string, error) {
	name, _, _, _, _, _, err := p.fillFromStatus()
	if err != nil {
		return "", err
	}
	return name, nil
}
func (p *Process) Exe() (string, error) {
	return p.fillFromExe()
}
func (p *Process) Cmdline() (string, error) {
	return p.fillFromCmdline()
}
func (p *Process) CreateTime() (int64, error) {
	_, _, _, createTime, _, err := p.fillFromStat()
	if err != nil {
		return 0, err
	}
	return createTime, nil
}

func (p *Process) Cwd() (string, error) {
	return p.fillFromCwd()
}
func (p *Process) Parent() (*Process, error) {
	return nil, NotImplementedError
}
func (p *Process) Status() (string, error) {
	_, status, _, _, _, _, err := p.fillFromStatus()
	if err != nil {
		return "", err
	}
	return status, nil
}
func (p *Process) Username() (string, error) {
	return "", nil
}
func (p *Process) Uids() ([]int32, error) {
	_, _, uids, _, _, _, err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	_, _, _, gids, _, _, err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	terminal, _, _, _, _, err := p.fillFromStat()
	if err != nil {
		return "", err
	}
	return terminal, nil
}
func (p *Process) Nice() (int32, error) {
	_, _, _, _, nice, err := p.fillFromStat()
	if err != nil {
		return 0, err
	}
	return nice, nil
}
func (p *Process) IOnice() (int32, error) {
	return 0, NotImplementedError
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	return nil, NotImplementedError
}

// FIXME: not sure if the func name is ok
func (p *Process) NetIOCounters() ([]NetIOCountersStat, error) {
	return fromNetDevFile(fmt.Sprintf("/proc/%d/net/dev", p.Pid))
}

func (p *Process) IOCounters() (*IOCountersStat, error) {
	return p.fillFromIO()
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	_, _, _, _, _, numCtxSwitches, err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	return numCtxSwitches, nil
}
func (p *Process) NumFDs() (int32, error) {
	return 0, NotImplementedError
}
func (p *Process) NumThreads() (int32, error) {
	_, _, _, _, numThreads, _, err := p.fillFromStatus()
	if err != nil {
		return 0, err
	}
	return numThreads, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, nil
}
func (p *Process) CPUTimes() (*CPUTimesStat, error) {
	_, _, cpuTimes, _, _, err := p.fillFromStat()
	if err != nil {
		return nil, err
	}
	return cpuTimes, nil
}
func (p *Process) CPUPpercent() (int32, error) {
	return 0, NotImplementedError
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, NotImplementedError
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	memInfo, _, err := p.fillFromStatm()
	if err != nil {
		return nil, err
	}
	return memInfo, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	_, memInfoEx, err := p.fillFromStatm()
	if err != nil {
		return nil, err
	}
	return memInfoEx, nil
}
func (p *Process) MemoryPercent() (float32, error) {
	return 0, NotImplementedError
}

func (p *Process) Children() ([]*Process, error) {
	return nil, NotImplementedError
}

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return nil, NotImplementedError
}

func (p *Process) Connections() ([]NetConnectionStat, error) {
	return nil, NotImplementedError
}

func (p *Process) IsRunning() (bool, error) {
	return true, NotImplementedError
}

// MemoryMaps get memory maps from /proc/(pid)/smaps
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	pid := p.Pid
	var ret []MemoryMapsStat
	smapsPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "smaps")
	contents, err := ioutil.ReadFile(smapsPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")

	// function of parsing a block
	getBlock := func(first_line []string, block []string) MemoryMapsStat {
		m := MemoryMapsStat{}
		m.Path = first_line[len(first_line)-1]

		for _, line := range block {
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			v := strings.Trim(field[1], " kB") // remove last "kB"
			switch field[0] {
			case "Size":
				m.Size = mustParseUint64(v)
			case "Rss":
				m.Rss = mustParseUint64(v)
			case "Pss":
				m.Pss = mustParseUint64(v)
			case "Shared_Clean":
				m.SharedClean = mustParseUint64(v)
			case "Shared_Dirty":
				m.SharedDirty = mustParseUint64(v)
			case "Private_Clean":
				m.PrivateClean = mustParseUint64(v)
			case "Private_Dirty":
				m.PrivateDirty = mustParseUint64(v)
			case "Referenced":
				m.Referenced = mustParseUint64(v)
			case "Anonymous":
				m.Anonymous = mustParseUint64(v)
			case "Swap":
				m.Swap = mustParseUint64(v)
			}
		}
		return m
	}

	blocks := make([]string, 16)
	for _, line := range lines {
		field := strings.Split(line, " ")
		if strings.HasSuffix(field[0], ":") == false {
			// new block section
			if len(blocks) > 0 {
				ret = append(ret, getBlock(field, blocks))
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

// Get num_fds from /proc/(pid)/fd
func (p *Process) fillFromfd() (int32, []*OpenFilesStat, error) {
	pid := p.Pid
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "fd")
	d, err := os.Open(statPath)
	if err != nil {
		return 0, nil, err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	numFDs := int32(len(fnames))

	openfiles := make([]*OpenFilesStat, numFDs)
	for _, fd := range fnames {
		fpath := filepath.Join(statPath, fd)
		filepath, err := os.Readlink(fpath)
		if err != nil {
			continue
		}
		o := &OpenFilesStat{
			Path: filepath,
			Fd:   mustParseUint64(fd),
		}
		openfiles = append(openfiles, o)
	}

	return numFDs, openfiles, nil
}

// Get cwd from /proc/(pid)/cwd
func (p *Process) fillFromCwd() (string, error) {
	pid := p.Pid
	cwdPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "cwd")
	cwd, err := os.Readlink(cwdPath)
	if err != nil {
		return "", err
	}
	return string(cwd), nil
}

// Get exe from /proc/(pid)/exe
func (p *Process) fillFromExe() (string, error) {
	pid := p.Pid
	exePath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "exe")
	exe, err := os.Readlink(exePath)
	if err != nil {
		return "", err
	}
	return string(exe), nil
}

// Get cmdline from /proc/(pid)/cmdline
func (p *Process) fillFromCmdline() (string, error) {
	pid := p.Pid
	cmdPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return "", err
	}
	// remove \u0000
	ret := strings.TrimFunc(string(cmdline), func(r rune) bool {
		if r == '\u0000' {
			return true
		}
		return false
	})

	return ret, nil
}

// Get IO status from /proc/(pid)/io
func (p *Process) fillFromIO() (*IOCountersStat, error) {
	pid := p.Pid
	ioPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "io")
	ioline, err := ioutil.ReadFile(ioPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(ioline), "\n")
	ret := &IOCountersStat{}

	for _, line := range lines {
		field := strings.Split(line, ":")
		if len(field) < 2 {
			continue
		}
		switch field[0] {
		case "rchar":
			ret.ReadCount = mustParseInt32(strings.Trim(field[1], " \t"))
		case "wchar":
			ret.WriteCount = mustParseInt32(strings.Trim(field[1], " \t"))
		case "read_bytes":
			ret.ReadBytes = mustParseInt32(strings.Trim(field[1], " \t"))
		case "write_bytes":
			ret.WriteBytes = mustParseInt32(strings.Trim(field[1], " \t"))
		}
	}

	return ret, nil
}

// Get memory info from /proc/(pid)/statm
func (p *Process) fillFromStatm() (*MemoryInfoStat, *MemoryInfoExStat, error) {
	pid := p.Pid
	memPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "statm")
	contents, err := ioutil.ReadFile(memPath)
	if err != nil {
		return nil, nil, err
	}
	fields := strings.Split(string(contents), " ")

	rss := mustParseUint64(fields[0]) * PageSize
	vms := mustParseUint64(fields[1]) * PageSize
	memInfo := &MemoryInfoStat{
		RSS: rss,
		VMS: vms,
	}
	memInfoEx := &MemoryInfoExStat{
		RSS:    rss,
		VMS:    vms,
		Shared: mustParseUint64(fields[2]) * PageSize,
		Text:   mustParseUint64(fields[3]) * PageSize,
		Lib:    mustParseUint64(fields[4]) * PageSize,
		Dirty:  mustParseUint64(fields[5]) * PageSize,
	}

	return memInfo, memInfoEx, nil
}

// Get various status from /proc/(pid)/status
func (p *Process) fillFromStatus() (string, string, []int32, []int32, int32, *NumCtxSwitchesStat, error) {
	pid := p.Pid
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "status")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return "", "", nil, nil, 0, nil, err
	}
	lines := strings.Split(string(contents), "\n")

	name := ""
	status := ""
	var numThreads int32
	var vol int32
	var unvol int32
	var uids []int32
	var gids []int32
	for _, line := range lines {
		field := strings.Split(line, ":")
		if len(field) < 2 {
			continue
		}
		//		fmt.Printf("%s ->__%s__\n", field[0], strings.Trim(field[1], " \t"))
		switch field[0] {
		case "Name":
			name = strings.Trim(field[1], " \t")
		case "State":
			// get between "(" and ")"
			s := strings.Index(field[1], "(") + 1
			e := strings.Index(field[1], "(") + 1
			status = field[1][s:e]
			//		case "PPid":  // filled by fillFromStat
		case "Uid":
			for _, i := range strings.Split(field[1], "\t") {
				uids = append(uids, mustParseInt32(i))
			}
		case "Gid":
			for _, i := range strings.Split(field[1], "\t") {
				gids = append(gids, mustParseInt32(i))
			}
		case "Threads":
			numThreads = mustParseInt32(field[1])
		case "voluntary_ctxt_switches":
			vol = mustParseInt32(field[1])
		case "nonvoluntary_ctxt_switches":
			unvol = mustParseInt32(field[1])
		}
	}

	numctx := &NumCtxSwitchesStat{
		Voluntary:   vol,
		Involuntary: unvol,
	}
	return name, status, uids, gids, numThreads, numctx, nil
}

func (p *Process) fillFromStat() (string, int32, *CPUTimesStat, int64, int32, error) {
	pid := p.Pid
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "stat")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return "", 0, nil, 0, 0, err
	}
	fields := strings.Fields(string(contents))

	termmap, err := getTerminalMap()
	terminal := ""
	if err == nil {
		terminal = termmap[mustParseUint64(fields[6])]
	}

	ppid := mustParseInt32(fields[3])
	utime, _ := strconv.ParseFloat(fields[13], 64)
	stime, _ := strconv.ParseFloat(fields[14], 64)

	cpuTimes := &CPUTimesStat{
		CPU:    "cpu",
		User:   float32(utime * (1000 / ClockTicks)),
		System: float32(stime * (1000 / ClockTicks)),
	}

	bootTime, _ := BootTime()
	ctime := ((mustParseUint64(fields[21]) / uint64(ClockTicks)) + uint64(bootTime)) * 1000
	createTime := int64(ctime)

	//	p.Nice = mustParseInt32(fields[18])
	// use syscall instead of parse Stat file
	snice, _ := syscall.Getpriority(PrioProcess, int(pid))
	nice := int32(snice) // FIXME: is this true?

	return terminal, ppid, cpuTimes, createTime, nice, nil
}

func Pids() ([]int32, error) {
	var ret []int32

	d, err := os.Open("/proc")
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
