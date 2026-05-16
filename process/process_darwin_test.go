// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package process

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumFDs(t *testing.T) {
	pid := os.Getpid()
	p, err := NewProcess(int32(pid))
	require.NoError(t, err)

	ctx := context.Background()

	before, err := p.NumFDsWithContext(ctx)
	require.NoError(t, err)

	// Open files to increase FD count
	f1, err := os.Open("/dev/null")
	require.NoError(t, err)
	defer f1.Close()

	f2, err := os.Open("/dev/null")
	require.NoError(t, err)
	defer f2.Close()

	after, err := p.NumFDsWithContext(ctx)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, after, before+2)
}

func TestNumFDs_NonExistent(t *testing.T) {
	p := &Process{Pid: 99999}
	_, err := p.NumFDsWithContext(context.Background())
	assert.Error(t, err)
}

func TestParseCmdline(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		nargs int
		want  []string
	}{
		{
			name:  "normal argv stops before envp",
			input: []byte("/bin/sh\x00\x00sh\x00-c\x00echo hi\x00HOME=/root\x00PATH=/bin\x00"),
			nargs: 3,
			want:  []string{"sh", "-c", "echo hi"},
		},
		{
			name:  "empty argv element does not leak envp",
			input: []byte("/bin/sh\x00\x00sh\x00\x00after\x00SECRET=hunter2\x00"),
			nargs: 3,
			want:  []string{"sh", "", "after"},
		},
		{
			name:  "multiple padding NULs between exec_path and argv",
			input: []byte("/usr/bin/prog\x00\x00\x00\x00\x00prog\x00arg1\x00ENV=value\x00"),
			nargs: 2,
			want:  []string{"prog", "arg1"},
		},
		{
			name:  "trailing empty argv element preserved",
			input: []byte("/bin/cmd\x00\x00cmd\x00arg\x00\x00ENV=x\x00"),
			nargs: 3,
			want:  []string{"cmd", "arg", ""},
		},
		{
			name:  "no env present",
			input: []byte("/bin/cmd\x00\x00cmd\x00only\x00"),
			nargs: 2,
			want:  []string{"cmd", "only"},
		},
		{
			name:  "nargs larger than available chunks does not panic",
			input: []byte("/bin/cmd\x00\x00cmd\x00arg\x00"),
			nargs: 99,
			want:  []string{"cmd", "arg", ""},
		},
		{
			name:  "zero nargs returns empty slice",
			input: []byte("/bin/cmd\x00\x00cmd\x00arg\x00"),
			nargs: 0,
			want:  []string{},
		},
		{
			name:  "empty buffer returns nil",
			input: []byte{},
			nargs: 0,
			want:  nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseCmdline(tc.input, tc.nargs)
			assert.Equal(t, tc.want, got)
		})
	}
}

func BenchmarkNumFDs(b *testing.B) {
	pid := int32(os.Getpid())
	p, err := NewProcess(pid)
	if err != nil {
		b.Skip(err)
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := p.NumFDsWithContext(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
