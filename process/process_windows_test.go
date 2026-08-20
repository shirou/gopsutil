package process

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestParseCommandLineInformation(t *testing.T) {
	headerSize := int(unsafe.Sizeof(windows.NTUnicodeString{}))

	buildBuf := func(s string) []byte {
		utf16Data := utf16.Encode([]rune(s))
		dataBytes := make([]byte, len(utf16Data)*2)
		for i, u := range utf16Data {
			binary.LittleEndian.PutUint16(dataBytes[i*2:], u)
		}
		buf := make([]byte, headerSize+len(dataBytes))
		binary.LittleEndian.PutUint16(buf[0:2], uint16(len(dataBytes))) // Length field
		copy(buf[headerSize:], dataBytes)
		return buf
	}

	tests := []struct {
		name      string
		buf       []byte
		expect    string
		expectErr bool
	}{
		{name: "valid command line", buf: buildBuf("notepad.exe file.txt"), expect: "notepad.exe file.txt"},
		{name: "empty command line", buf: buildBuf(""), expect: ""},
		{name: "buffer too small for length field", buf: []byte{0x01}, expectErr: true},
		{
			name: "truncated buffer (length claims more than available)",
			buf: func() []byte {
				b := buildBuf("some command")
				return b[:len(b)-4] // simulates the panic scenario point 4 guards against
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommandLineInformation(tt.buf)
			if (err != nil) != tt.expectErr {
				t.Fatalf("unexpected error state: err=%v, expectErr=%v", err, tt.expectErr)
			}
			if !tt.expectErr && got != tt.expect {
				t.Errorf("got %q, expect %q", got, tt.expect)
			}
		})
	}
}
