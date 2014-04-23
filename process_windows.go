// +build windows

package gopsutil

func Pids() ([]int32, error) {
	ret := make([]int32, 0)

	return ret, nil
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}
	return p, nil
}
