// SPDX-License-Identifier: BSD-3-Clause
package process

import (
	"encoding/binary"
	"os"
	"syscall"
	"testing"
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

	nativeCmdline, err := getProcessCommandLineNative(h, pid)
	require.NoError(t, err)
	assert.Equal(t, pebCmdline, nativeCmdline)
}
