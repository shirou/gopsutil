// SPDX-License-Identifier: BSD-3-Clause
//go:build freebsd || darwin

package net

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lsof escapes an ASCII space in COMMAND as \x20 but, in a UTF-8 locale,
// prints other Unicode spaces such as U+00A0 as is.
func TestParseNetLineUnicodeSpaceCommand(t *testing.T) {
	line := "x\u00a0y      24696 user    3u  IPv4 0xe4d6d32d671ee13b      0t0  TCP 127.0.0.1:45680->127.0.0.1:50000 (ESTABLISHED)"

	n, err := parseNetLine(line)
	require.NoError(t, err)
	assert.Equal(t, int32(24696), n.Pid)
	assert.Equal(t, uint32(3), n.Fd)
	assert.Equal(t, uint32(syscall.AF_INET), n.Family)
	assert.Equal(t, uint32(syscall.SOCK_STREAM), n.Type)
	assert.Equal(t, Addr{IP: "127.0.0.1", Port: 45680}, n.Laddr)
	assert.Equal(t, Addr{IP: "127.0.0.1", Port: 50000}, n.Raddr)
	assert.Equal(t, "ESTABLISHED", n.Status)
}
