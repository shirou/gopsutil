// SPDX-License-Identifier: BSD-3-Clause
//go:build linux || freebsd

package process

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func Test_SendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := NewProcess(int32(checkPid))
	assert.NoErrorf(t, p.SendSignal(unix.SIGCONT), "send signal")
}
