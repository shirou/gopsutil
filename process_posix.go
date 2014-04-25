// +build linux freebsd

package gopsutil

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// POSIX
func getTerminalMap() (map[uint64]string, error) {
	ret := make(map[uint64]string)
	termfiles := make([]string, 0)

	d, err := os.Open("/dev")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	devnames, err := d.Readdirnames(-1)
	for _, devname := range devnames {
		if strings.HasPrefix(devname, "/dev/tty") {
			termfiles = append(termfiles, "/dev/tty/"+devname)
		}
	}

	ptsd, err := os.Open("/dev/pts")
	if err != nil {
		return nil, err
	}
	defer ptsd.Close()

	ptsnames, err := ptsd.Readdirnames(-1)
	for _, ptsname := range ptsnames {
		termfiles = append(termfiles, "/dev/pts/"+ptsname)
	}

	for _, name := range termfiles {
		stat := syscall.Stat_t{}
		syscall.Stat(name, &stat)
		rdev := uint64(stat.Rdev)
		ret[rdev] = strings.Replace(name, "/dev", "", -1)
	}
	return ret, nil
}

func (p *Process) Send_signal(sig syscall.Signal) error {
	sig_as_str := "INT"
	switch sig {
	case syscall.SIGSTOP:
		sig_as_str = "STOP"
	case syscall.SIGCONT:
		sig_as_str = "CONT"
	case syscall.SIGTERM:
		sig_as_str = "TERM"
	case syscall.SIGKILL:
		sig_as_str = "KILL"
	}

	cmd := exec.Command("kill", "-s", sig_as_str, strconv.Itoa(int(p.Pid)))
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (p *Process) Suspend() error {
	return p.Send_signal(syscall.SIGSTOP)
}
func (p *Process) Resume() error {
	return p.Send_signal(syscall.SIGCONT)
}
func (p *Process) Terminate() error {
	return p.Send_signal(syscall.SIGTERM)
}
func (p *Process) Kill() error {
	return p.Send_signal(syscall.SIGKILL)
}
