// +build linux freebsd

package gopsutil

import (
	"os"
	"syscall"
	"testing"
)

func TestSendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := NewProcess(int32(checkPid))
	err := p.SendSignal(syscall.SIGCONT)
	if err != nil {
		t.Errorf("send signal %v", err)
	}
}
