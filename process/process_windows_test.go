// SPDX-License-Identifier: BSD-3-Clause
package process

import (
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestParseCommandLineInformation(t *testing.T) {
	ntUnicodeStringHeaderSize := int(unsafe.Offsetof(windows.NTUnicodeString{}.Buffer)) +
		int(unsafe.Sizeof((*uint16)(nil)))

	buildCommandLineInfoBuf := func(s string) []byte {
		utf16Data := utf16.Encode([]rune(s))
		dataBytes := make([]byte, len(utf16Data)*2)
		for i, u := range utf16Data {
			binary.LittleEndian.PutUint16(dataBytes[i*2:], u)
		}
		buf := make([]byte, ntUnicodeStringHeaderSize+len(dataBytes))
		binary.LittleEndian.PutUint16(buf[0:2], uint16(len(dataBytes)))
		copy(buf[ntUnicodeStringHeaderSize:], dataBytes)
		return buf
	}
	tests := []struct {
		name      string
		buf       []byte
		expect    string
		expectErr bool
	}{
		{
			name:   "valid command line",
			buf:    buildCommandLineInfoBuf("notepad.exe file.txt"),
			expect: "notepad.exe file.txt",
		},
		{
			name:   "empty command line",
			buf:    buildCommandLineInfoBuf(""),
			expect: "",
		},
		{
			// The kernel sizes the buffer by returnLength, which is not
			// necessarily header + Length, so anything past Length has to be
			// ignored. The filler is non-zero on purpose: a slice that ran to
			// the end of the buffer would otherwise decode to the same string.
			name: "buffer padded past the declared length",
			buf: func() []byte {
				b := buildCommandLineInfoBuf("notepad.exe file.txt")
				return append(b, 0xAA, 0xBB, 0xCC, 0xDD)
			}(),
			expect: "notepad.exe file.txt",
		},
		{
			name:      "buffer too small for length field",
			buf:       []byte{0x01},
			expectErr: true,
		},
		{
			name: "truncated buffer (length claims more than available)",
			buf: func() []byte {
				b := buildCommandLineInfoBuf("some command")
				return b[:len(b)-4]
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommandLineInformation(tt.buf)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expect, got)
		})
	}
}

// TestGetProcessCommandLineNativeMatchesPEB verifies that the native
// NtQueryInformationProcess path and the PEB read path return the same
// command line for the current process.
func TestGetProcessCommandLineNativeMatchesPEB(t *testing.T) {
	pid := int32(os.Getpid())
	h, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
		false,
		uint32(pid),
	)
	require.NoError(t, err)
	defer syscall.CloseHandle(syscall.Handle(h))

	pebCmdline, err := getProcessCommandLinePEB(h)
	require.NoError(t, err)
	// A test binary always has a command line. Without this, a regression that
	// made both paths return "" (a Length read at the wrong offset, say) would
	// still satisfy the comparison below.
	require.NotEmpty(t, pebCmdline)

	nativeCmdline, err := getProcessCommandLineNative(h, pid)
	require.NoError(t, err)
	assert.Equal(t, pebCmdline, nativeCmdline)
}

// TestKillReusedPid checks that a Process whose PID now belongs to another
// process (simulated with a create time that does not match) is not killed.
func TestKillReusedPid(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "ping", "localhost", "-n", "30")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
	})
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	p, err := NewProcess(int32(cmd.Process.Pid))
	require.NoError(t, err)
	createTime, err := p.CreateTime()
	require.NoError(t, err)

	stale := &Process{Pid: p.Pid, createTime: createTime - 1000}
	require.ErrorIs(t, stale.Kill(), ErrorProcessNotRunning)
	require.ErrorIs(t, stale.Terminate(), ErrorProcessNotRunning)
	require.ErrorIs(t, stale.Suspend(), ErrorProcessNotRunning)
	require.ErrorIs(t, stale.Resume(), ErrorProcessNotRunning)
	select {
	case <-done:
		t.Fatal("process was killed through a stale Process")
	case <-time.After(500 * time.Millisecond):
	}

	require.NoError(t, p.Kill())
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("process was not killed")
	}
}
