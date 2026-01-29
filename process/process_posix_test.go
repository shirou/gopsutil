// SPDX-License-Identifier: BSD-3-Clause
//go:build linux || freebsd

package process

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func Test_SendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := NewProcess(int32(checkPid))
	assert.NoErrorf(t, p.SendSignal(unix.SIGCONT), "send signal")
}

func TestGetTerminalMapPathsExist(t *testing.T) {
	termmap, err := getTerminalMap()
	if err != nil {
		t.Skipf("getTerminalMap not available: %v", err)
	}

	for _, name := range termmap {
		fullPath := common.HostDev(name)
		_, err := os.Stat(fullPath)
		assert.NoErrorf(t, err, "terminal device should exist: %s", fullPath)
	}
}
