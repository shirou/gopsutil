// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package net

import (
	"context"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample netstat -Aan output captured from AIX 7.3 for unit testing.
const testNetstatAanInet = `Active Internet connections (including servers)
PCB/ADDR         Proto Recv-Q Send-Q  Local Address      Foreign Address    (state)
f1000f00055bcbc0 tcp        0      0  *.*                   *.*                   CLOSED
f1000f00055963c0 tcp4       0      0  *.*                   *.*                   CLOSED
f1000f000017e3c0 tcp6       0      0  *.22                  *.*                   LISTEN
f1000f0000287bc0 tcp4       0      0  *.22                  *.*                   LISTEN
f1000f000017f3c0 tcp4       0      0  *.25                  *.*                   LISTEN
f1000f00002d8bc0 tcp4       0      0  192.168.242.122.22    24.236.207.124.33236  ESTABLISHED
f1000f00055eabc0 tcp6       0      0  ::1.6010              *.*                   LISTEN
f1000f000561c3c0 tcp4       0      0  192.168.242.122.34898 34.120.255.184.443    ESTABLISHED
f1000f0000140600 udp        0      0  *.111                 *.*
f1000f000017be00 udp        0      0  *.161                 *.*
f1000f0005582e00 udp        0      0  *.514                 *.*
`

const testNetstatAanUnix = `Active UNIX domain sockets
SADR/PCB         Type   Recv-Q Send-Q      Inode            Conn             Refs           Nextref      Addr
f1000f0000144808 dgram       0      0                0 f1000f00055a1b00                0                0
f1000f0000145580
f1000f000010a408 dgram       0      0 f1000b02a0354c20                0                0                0 /dev/SRC
f1000f000557b280
f1000f0000169808 stream      0      0 f1000b02a03a9420                0                0                0 /var/ct/IW/soc/mc/RMIBM.DRM.0
f1000f0000177780
f1000f000015c808 stream      0      0                0 f1000f0000177580                0                0
f1000f0000177880
`

const testNetstatAanFull = testNetstatAanInet + "\n" + testNetstatAanUnix

func TestParseNetstatAanInet(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanInet, "inet")
	require.NoError(t, err)

	// Should skip *.*  local addresses (2 CLOSED entries) and include the rest
	// Expected: *.22 tcp6, *.22 tcp4, *.25 tcp4, ESTABLISHED ssh, ::1.6010, ESTABLISHED https,
	//           *.111 udp, *.161 udp, *.514 udp = 9 entries
	assert.Len(t, entries, 9)

	// Verify the ESTABLISHED TCP connection
	var sshConn *netstatAanEntry
	for i := range entries {
		if entries[i].conn.Status == "ESTABLISHED" && entries[i].conn.Laddr.Port == 22 {
			sshConn = &entries[i]
			break
		}
	}
	require.NotNil(t, sshConn, "should find the SSH ESTABLISHED connection")
	assert.Equal(t, "f1000f00002d8bc0", sshConn.sockAddr)
	assert.Equal(t, "tcp4", sshConn.proto)
	assert.Equal(t, uint32(syscall.AF_INET), sshConn.conn.Family)
	assert.Equal(t, uint32(syscall.SOCK_STREAM), sshConn.conn.Type)
	assert.Equal(t, "192.168.242.122", sshConn.conn.Laddr.IP)
	assert.Equal(t, uint32(22), sshConn.conn.Laddr.Port)
	assert.Equal(t, "24.236.207.124", sshConn.conn.Raddr.IP)
	assert.Equal(t, uint32(33236), sshConn.conn.Raddr.Port)

	// Verify a LISTEN entry
	var listenConn *netstatAanEntry
	for i := range entries {
		if entries[i].conn.Status == "LISTEN" && entries[i].conn.Laddr.Port == 25 {
			listenConn = &entries[i]
			break
		}
	}
	require.NotNil(t, listenConn, "should find the SMTP LISTEN connection")
	assert.Equal(t, "0.0.0.0", listenConn.conn.Laddr.IP)
	assert.Equal(t, uint32(25), listenConn.conn.Laddr.Port)

	// Verify IPv6 LISTEN entry
	var ipv6Conn *netstatAanEntry
	for i := range entries {
		if entries[i].proto == "tcp6" && entries[i].conn.Laddr.Port == 22 {
			ipv6Conn = &entries[i]
			break
		}
	}
	require.NotNil(t, ipv6Conn, "should find the IPv6 SSH LISTEN connection")
	assert.Equal(t, uint32(syscall.AF_INET6), ipv6Conn.conn.Family)
	assert.Equal(t, "::", ipv6Conn.conn.Laddr.IP)

	// Verify UDP entry (no state field)
	var udpConn *netstatAanEntry
	for i := range entries {
		if entries[i].conn.Laddr.Port == 161 {
			udpConn = &entries[i]
			break
		}
	}
	require.NotNil(t, udpConn, "should find the SNMP UDP connection")
	assert.Equal(t, uint32(syscall.SOCK_DGRAM), udpConn.conn.Type)
	assert.Empty(t, udpConn.conn.Status)
}

func TestParseNetstatAanTCPFilter(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanInet, "tcp")
	require.NoError(t, err)

	for _, entry := range entries {
		assert.True(t, strings.HasPrefix(entry.proto, "tcp"),
			"tcp filter should only return TCP entries, got proto=%s", entry.proto)
	}
}

func TestParseNetstatAanUDPFilter(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanInet, "udp")
	require.NoError(t, err)

	assert.Len(t, entries, 3)
	for _, entry := range entries {
		assert.True(t, strings.HasPrefix(entry.proto, "udp"),
			"udp filter should only return UDP entries, got proto=%s", entry.proto)
	}
}

func TestParseNetstatAanUnix(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanUnix, "unix")
	require.NoError(t, err)

	// 4 unix socket entries (each spans 2 lines; second line is skipped)
	assert.Len(t, entries, 4)

	// Verify dgram entry with address
	var srcConn *netstatAanEntry
	for i := range entries {
		if entries[i].conn.Laddr.IP == "/dev/SRC" {
			srcConn = &entries[i]
			break
		}
	}
	require.NotNil(t, srcConn, "should find the /dev/SRC unix socket")
	assert.Equal(t, uint32(syscall.AF_UNIX), srcConn.conn.Family)
	assert.Equal(t, uint32(syscall.SOCK_DGRAM), srcConn.conn.Type)

	// Verify stream entry
	var streamConn *netstatAanEntry
	for i := range entries {
		if entries[i].conn.Type == uint32(syscall.SOCK_STREAM) {
			streamConn = &entries[i]
			break
		}
	}
	require.NotNil(t, streamConn, "should find a stream unix socket")
	assert.Equal(t, uint32(syscall.AF_UNIX), streamConn.conn.Family)
}

func TestParseNetstatAanAll(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanFull, "all")
	require.NoError(t, err)

	// Should include both inet and unix entries
	var hasInet, hasUnix bool
	for _, entry := range entries {
		if entry.proto == "unix" {
			hasUnix = true
		} else {
			hasInet = true
		}
	}
	assert.True(t, hasInet, "should have inet entries")
	assert.True(t, hasUnix, "should have unix entries")
}

func TestParseNetstatAanUnixFilterExcludesInet(t *testing.T) {
	entries, err := parseNetstatAan(testNetstatAanFull, "unix")
	require.NoError(t, err)

	for _, entry := range entries {
		assert.Equal(t, "unix", entry.proto,
			"unix filter should exclude inet entries")
	}
}

func TestParseAIXRmsockPid(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int32
	}{
		{
			// AIX has a known typo: double 'c' in "process"
			name:   "AIX typo double c",
			output: "The socket 0xf1000f00002d8808 is being held by proc" + "cess 21496092 (sshd).",
			want:   21496092,
		},
		{
			// Handle correct spelling too in case a future AIX version fixes it
			name:   "correct spelling process",
			output: "The socket 0xf1000f00002d8808 is being held by process 14287304 (sshd).",
			want:   14287304,
		},
		{
			name:   "wait for exiting processes",
			output: "Wait for exiting processes to be cleaned up before removing the socket",
			want:   0,
		},
		{
			name:   "not a socket",
			output: "It is not a socket",
			want:   0,
		},
		{
			name:   "kernel address error",
			output: "rmsock : Unable to read kernel address f1000f0000140600, errno = 22",
			want:   0,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAIXRmsockPid(tt.output)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Integration tests below require running on AIX with real netstat/rmsock.

func TestConnectionsPidWithContext(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	conns, err := ConnectionsPidWithContext(ctx, "inet", pid)
	if err != nil {
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
		for _, conn := range conns {
			assert.NotEmpty(t, conn.Family)
			assert.NotEmpty(t, conn.Type)
		}
	}
}

func TestConnectionsPidWithContextAll(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	conns, err := ConnectionsPidWithContext(ctx, "all", pid)
	if err != nil {
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
	}
}

func TestConnectionsPidWithContextUDP(t *testing.T) {
	ctx := context.Background()
	pid := int32(os.Getpid())

	conns, err := ConnectionsPidWithContext(ctx, "udp", pid)
	if err != nil {
		t.Logf("ConnectionsPidWithContext error: %v", err)
		return
	}
	if conns != nil {
		assert.IsType(t, []ConnectionStat{}, conns)
	}
}

func TestConnectionsWithContext(t *testing.T) {
	ctx := context.Background()

	conns, err := ConnectionsWithContext(ctx, "inet")
	require.NoError(t, err)
	assert.NotNil(t, conns)
	assert.IsType(t, []ConnectionStat{}, conns)
	assert.NotEmpty(t, conns)

	for _, conn := range conns {
		assert.NotEmpty(t, conn.Family)
		assert.NotEmpty(t, conn.Type)
	}
}
