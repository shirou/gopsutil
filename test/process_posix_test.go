// +build linux freebsd

package test

import (
	"os"
	"syscall"
	"testing"

	"github.com/shirou/gopsutil"
)

func Test_SendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := gopsutil.NewProcess(int32(checkPid))
	err := p.SendSignal(syscall.SIGCONT)
	if err != nil {
		t.Errorf("send signal %v", err)
	}
}
