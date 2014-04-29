// +build linux

package gopsutil

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	PRIO_PROCESS = 0 // linux/resource.h
)

// Memory_info_ex is different between OSes
type Memory_info_exStat struct {
	RSS    uint64 `json:"rss"`    // bytes
	VMS    uint64 `json:"vms"`    // bytes
	Shared uint64 `json:"shared"` // bytes
	Text   uint64 `json:"text"`   // bytes
	Lib    uint64 `json:"lib"`    // bytes
	Data   uint64 `json:"data"`   // bytes
	Dirty  uint64 `json:"dirty"`  // bytes
}

type Memory_mapsStat struct {
	Path          string `json:"path"`
	Rss           uint64 `json:"rss"`
	Size          uint64 `json:"size"`
	Pss           uint64 `json:"pss"`
	Shared_clean  uint64 `json:"shared_clean"`
	Shared_dirty  uint64 `json:"shared_dirty"`
	Private_clean uint64 `json:"private_clean"`
	Private_dirty uint64 `json:"private_dirty"`
	Referenced    uint64 `json:"referenced"`
	Anonymous     uint64 `json:"anonymous"`
	Swap          uint64 `json:"swap"`
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
	name, _, _, _, _, err := p.fillFromStatus()
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
func (p *Process) Cwd() (string, error) {
	return p.fillFromCwd()
}
func (p *Process) Parent() (*Process, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Status() (string, error) {
	_, status, _, _, _, err := p.fillFromStatus()
	if err != nil {
		return "", err
	}
	return status, nil
}
func (p *Process) Username() (string, error) {
	return "", nil
}
func (p *Process) Uids() ([]int32, error) {
	_, _, uids, _, _, err := p.fillFromStatus()
	if err != nil {
		return nil, err
	}
	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	_, _, _, gids, _, err := p.fillFromStatus()
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
func (p *Process) Ionice() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	return nil, errors.New("Not implemented yet")
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
	_, _, _, _, num_threads, err := p.fillFromStatus()
	if err != nil {
		return 0, err
	}
	return num_threads, nil
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, nil
}
func (p *Process) Cpu_times() (*CPU_TimesStat, error) {
	_, _, cpu_times, _, _, err := p.fillFromStat()
	if err != nil {
		return nil, err
	}
	return cpu_times, nil
}
func (p *Process) Cpu_percent() (int32, error) {
	return 0, errors.New("Not implemented yet")
}
func (p *Process) Cpu_affinity() ([]int32, error) {
	return nil, errors.New("Not implemented yet")
}
func (p *Process) Memory_info() (*Memory_infoStat, error) {
	mem_info, _, err := p.fillFromStatm()
	if err != nil {
		return nil, err
	}
	return mem_info, nil
}
func (p *Process) Memory_info_ex() (*Memory_info_exStat, error) {
	_, mem_info_ex, err := p.fillFromStatm()
	if err != nil {
		return nil, err
	}
	return mem_info_ex, nil
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

// Get memory maps from /proc/(pid)/smaps
func (p *Process) Memory_Maps() (*[]Memory_mapsStat, error) {
	pid := p.Pid
	ret := make([]Memory_mapsStat, 0)
	smapsPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "smaps")
	contents, err := ioutil.ReadFile(smapsPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")

	// function of parsing a block
	get_block := func(first_line []string, block []string) Memory_mapsStat {
		m := Memory_mapsStat{}
		m.Path = first_line[len(first_line)-1]

		for _, line := range block {
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			v := strings.Trim(field[1], " kB") // remove last "kB"
			switch field[0] {
			case "Size":
				m.Size = parseUint64(v)
			case "Rss":
				m.Rss = parseUint64(v)
			case "Pss":
				m.Pss = parseUint64(v)
			case "Shared_Clean":
				m.Shared_clean = parseUint64(v)
			case "Shared_Dirty":
				m.Shared_dirty = parseUint64(v)
			case "Private_Clean":
				m.Private_clean = parseUint64(v)
			case "Private_Dirty":
				m.Private_dirty = parseUint64(v)
			case "Referenced":
				m.Referenced = parseUint64(v)
			case "Anonymous":
				m.Anonymous = parseUint64(v)
			case "Swap":
				m.Swap = parseUint64(v)
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
				ret = append(ret, get_block(field, blocks))
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
func (p *Process) fillFromfd() (int32, []*Open_filesStat, error) {
	pid := p.Pid
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "fd")
	d, err := os.Open(statPath)
	if err != nil {
		return 0, nil, err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	num_fds := int32(len(fnames))

	openfiles := make([]*Open_filesStat, num_fds)
	for _, fd := range fnames {
		fpath := filepath.Join(statPath, fd)
		filepath, err := os.Readlink(fpath)
		if err != nil {
			continue
		}
		o := &Open_filesStat{
			Path: filepath,
			Fd:   parseUint64(fd),
		}
		openfiles = append(openfiles, o)
	}

	return num_fds, openfiles, nil
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

// Get memory info from /proc/(pid)/statm
func (p *Process) fillFromStatm() (*Memory_infoStat, *Memory_info_exStat, error) {
	pid := p.Pid
	memPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "statm")
	contents, err := ioutil.ReadFile(memPath)
	if err != nil {
		return nil, nil, err
	}
	fields := strings.Split(string(contents), " ")

	rss := parseUint64(fields[0]) * PAGESIZE
	vms := parseUint64(fields[1]) * PAGESIZE
	mem_info := &Memory_infoStat{
		RSS: rss,
		VMS: vms,
	}
	mem_info_ex := &Memory_info_exStat{
		RSS:    rss,
		VMS:    vms,
		Shared: parseUint64(fields[2]) * PAGESIZE,
		Text:   parseUint64(fields[3]) * PAGESIZE,
		Lib:    parseUint64(fields[4]) * PAGESIZE,
		Dirty:  parseUint64(fields[5]) * PAGESIZE,
	}

	return mem_info, mem_info_ex, nil
}

// Get various status from /proc/(pid)/status
func (p *Process) fillFromStatus() (string, string, []int32, []int32, int32, error) {
	pid := p.Pid
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "status")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return "", "", nil, nil, 0, err
	}
	lines := strings.Split(string(contents), "\n")

	name := ""
	status := ""
	var num_threads int32
	uids := make([]int32, 0)
	gids := make([]int32, 0)
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
				uids = append(uids, parseInt32(i))
			}
		case "Gid":
			for _, i := range strings.Split(field[1], "\t") {
				gids = append(gids, parseInt32(i))
			}
		case "Threads":
			num_threads = parseInt32(field[1])
		}
	}

	return name, status, uids, gids, num_threads, nil
}

func (p *Process) fillFromStat() (string, int32, *CPU_TimesStat, int64, int32, error) {
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
		terminal = termmap[parseUint64(fields[6])]
	}

	ppid := parseInt32(fields[3])
	utime, _ := strconv.ParseFloat(fields[13], 64)
	stime, _ := strconv.ParseFloat(fields[14], 64)

	cpu_times := &CPU_TimesStat{
		Cpu:    "cpu",
		User:   float32(utime * (1000 / CLOCK_TICKS)),
		System: float32(stime * (1000 / CLOCK_TICKS)),
	}

	boot_time, _ := Boot_time()
	ctime := ((parseUint64(fields[21]) / uint64(CLOCK_TICKS)) + uint64(boot_time)) * 1000
	create_time := int64(ctime)

	//	p.Nice = parseInt32(fields[18])
	// use syscall instead of parse Stat file
	snice, _ := syscall.Getpriority(PRIO_PROCESS, int(pid))
	nice := int32(snice) // FIXME: is this true?

	return terminal, ppid, cpu_times, create_time, nice, nil
}

func Pids() ([]int32, error) {
	ret := make([]int32, 0)

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
