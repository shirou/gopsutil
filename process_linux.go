// +build linux

package gopsutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewProcess(pid int32) (Process, error) {
	p := Process{
		Pid: int32(pid),
	}
	go fillFromStat(pid, &p)

	/*
	   //	user := parseInt32(fields[13])
	   	//sys := parseInt32(fields[14])
	   	// convert to millis
	   	self.User = user * (1000 / system.ticks)
	   	self.Sys = sys * (1000 / system.ticks)
	   	self.Total = self.User + self.Sys

	   	// convert to millis
	   	self.StartTime, _ = strtoull(fields[21])
	   	self.StartTime /= system.ticks
	   	self.StartTime += system.btime
	   	self.StartTime *= 1000
	*/
	return p, nil
}

func parseInt32(val string) int32 {
	vv, _ := strconv.ParseInt(val, 10, 32)
	return int32(vv)
}

/*
func fillFromStatm(pid int32, p *Process) error{
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "statm")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(contents))

}
*/

func fillFromStat(pid int32, p *Process) error {
	statPath := filepath.Join("/", "proc", strconv.Itoa(int(pid)), "stat")
	contents, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(contents))

	p.Name = strings.Trim(fields[1], "()") // remove "(" and ")"
	p.Status, _ = getState(fields[2][0])
	p.Ppid = parseInt32(fields[3])
	//	p.Terminal, _ = strconv.Atoi(fields[6])
	//	p.Priority, _ = strconv.Atoi(fields[17])
	p.Nice = parseInt32(fields[18])
	//	p.Processor, _ = strconv.Atoi(fields[38])
	return nil
}

func getState(status uint8) (string, error) {

	/*
	   >>> psutil.STATUS_RUNNING
	   'running'
	   >>> psutil.STATUS_SLEEPING
	   'sleeping'
	   >>> psutil.STATUS_DISK_SLEEP
	   'disk-sleep'
	   >>> psutil.STATUS_STOPPED
	   'stopped'
	   >>> psutil.STATUS_TRACING_STOP
	   'tracing-stop'
	   >>> psutil.STATUS_ZOMBIE
	   'zombie'
	   >>> psutil.STATUS_DEAD
	   'dead'
	   >>> psutil.STATUS_WAKE_KILL
	   Traceback (most recent call last):
	     File "<stdin>", line 1, in <module>
	   AttributeError: 'ModuleWrapper' object has no attribute 'STATUS_WAKE_KILL'
	   >>> psutil.STATUS_WAKING
	   'waking'
	   >>> psutil.STATUS_IDLE
	   'idle'
	   >>> psutil.STATUS_LOCKED
	   'locked'
	   >>> psutil.STATUS_WAITING
	   'waiting'
	*/
	return "running", nil
}

func processes() ([]Process, error) {
	ret := make([]Process, 0)

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
