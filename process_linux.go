// +build linux

package gopsutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	PRIO_PROCESS = 0 // linux/resource.h
)

type fillFunc func(pid int32, p *Process) error

func NewProcess(pid int32) (*Process, error) {
	p := &Process{
		Pid: int32(pid),
	}

	// Fill Process information from fillFuncs
	var wg sync.WaitGroup
	funcs := []fillFunc{fillFromStat, fillFromStatus, fillFromfd, fillFromCmdline}

	wg.Add(len(funcs))
	for _, f := range funcs {
		go func(f fillFunc) {
			wg.Done()
			f(pid, p)
		}(f)
	}
	wg.Wait()

	return p, nil
}

// Parse to int32 without error
func parseInt32(val string) int32 {
	vv, _ := strconv.ParseInt(val, 10, 32)
	return int32(vv)
}

// Parse to uint64 without error
func parseUint64(val string) uint64 {
	vv, _ := strconv.ParseInt(val, 10, 64)
	return uint64(vv)
}

// Get num_fds from /proc/(pid)/fd
func fillFromfd(pid int32, p *Process) error {
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "fd")
	d, err := os.Open(statPath)
	if err != nil {
		return err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	p.Num_fds = int32(len(fnames))

	return nil
}

// Get cmdline from /proc/(pid)/cmdline
func fillFromCmdline(pid int32, p *Process) error {
	cmdPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "cmdline")
	cmdline, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return err
	}
	// remove \u0000
	p.Cmdline = strings.TrimFunc(string(cmdline), func(r rune) bool {
		if r == '\u0000' {
			return true
		}
		return false
	})

	return nil
}

// get various status from /proc/(pid)/status
func fillFromStatus(pid int32, p *Process) error {
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "status")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contents), "\n")

	for _, line := range lines {
		field := strings.Split(line, ":")
		if len(field) < 2 {
			continue
		}
		//		fmt.Printf("%s ->__%s__\n", field[0], strings.Trim(field[1], " \t"))
		switch field[0] {
		case "Name":
			p.Name = strings.Trim(field[1], " \t")
		case "State":
			// get between "(" and ")"
			s := strings.Index(field[1], "(") + 1
			e := strings.Index(field[1], "(") + 1
			p.Status = field[1][s:e]
			//		case "PPid":  // filled by fillFromStat
		case "Uid":
			for _, i := range strings.Split(field[1], "\t") {
				p.Uids = append(p.Uids, parseInt32(i))
			}
		case "Gid":
			for _, i := range strings.Split(field[1], "\t") {
				p.Gids = append(p.Uids, parseInt32(i))
			}
		case "Threads":
			p.Num_Threads = parseInt32(field[1])
		}
	}

	return nil
}

func fillFromStat(pid int32, p *Process) error {
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "stat")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(contents))

	termmap, err := getTerminalMap()
	if err == nil {
		p.Terminal = termmap[parseUint64(fields[6])]
	}

	p.Ppid = parseInt32(fields[3])
	utime, _ := strconv.ParseFloat(fields[13], 64)
	stime, _ := strconv.ParseFloat(fields[14], 64)

	p.Cpu_times = CPU_TimesStat{
		Cpu:    "cpu",
		User:   float32(utime * (1000 / CLOCK_TICKS)),
		System: float32(stime * (1000 / CLOCK_TICKS)),
	}

	boot_time, _ := Boot_time()
	ctime := ((parseUint64(fields[21]) / uint64(CLOCK_TICKS)) + uint64(boot_time)) * 1000
	p.Create_time = int64(ctime)

	//	p.Nice = parseInt32(fields[18])
	// use syscall instead of parse Stat file
	nice, _ := syscall.Getpriority(PRIO_PROCESS, int(pid))
	p.Nice = int32(nice) // FIXME: is this true?

	return nil
}

func processes() ([]*Process, error) {
	ret := make([]*Process, 0)

	pids, err := Pids()
	if err != nil {
		return ret, err
	}

	for _, pid := range pids {
		p, err := NewProcess(pid)
		if err != nil {
			continue // FIXME: should return error?
		}
		ret = append(ret, p)
	}

	return ret, nil
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
