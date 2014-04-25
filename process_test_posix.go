// +build linux freebsd

package gopsutil

import (
	"os"
	"syscall"
	"testing"
)

func Test_SendSignal(t *testing.T) {
	check_pid := os.Getpid()

	p, _ := NewProcess(int32(check_pid))
	err := p.Send_signal(syscall.SIGCONT)
	if err != nil {
		t.Errorf("send signal %v", err)
	}
}
