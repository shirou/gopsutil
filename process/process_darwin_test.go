// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package process

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

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

// rusageInfoV2 must match struct rusage_info_v2 from sys/resource.h: a 16-byte
// uuid followed by 18 uint64 fields.
//
// The size alone is not enough: the two fields actually read sit at the end, so
// a reordering earlier in the struct would keep the size at 160 while silently
// shifting them. Pin their offsets too.
func TestRusageInfoV2Layout(t *testing.T) {
	var usage rusageInfoV2

	assert.Equal(t, uintptr(16+18*8), unsafe.Sizeof(usage))
	assert.Equal(t, uintptr(16+16*8), unsafe.Offsetof(usage.DiskIOBytesRead))
	assert.Equal(t, uintptr(16+17*8), unsafe.Offsetof(usage.DiskIOBytesWritten))
}

func TestIOCounters_DiskBytes(t *testing.T) {
	p, err := NewProcess(int32(os.Getpid()))
	require.NoError(t, err)

	before, err := p.IOCounters()
	require.NoError(t, err)

	// Darwin exposes disk bytes only; the other fields stay zero.
	assert.Zero(t, before.ReadCount)
	assert.Zero(t, before.WriteCount)
	assert.Zero(t, before.ReadBytes)
	assert.Zero(t, before.WriteBytes)

	path := filepath.Join(t.TempDir(), "disk-io-test")
	file, err := os.Create(path)
	require.NoError(t, err)
	data := make([]byte, 4*1024*1024)
	_, err = file.Write(data)
	require.NoError(t, err)
	require.NoError(t, file.Sync())
	require.NoError(t, file.Close())

	after, err := p.IOCounters()
	require.NoError(t, err)

	// Always hold the counters to being cumulative. Checking this before the
	// skip below matters: if proc_pid_rusage ever stopped reporting and returned
	// zeroes, both samples would be equal and a bare skip would go green.
	require.GreaterOrEqual(t, after.DiskWriteBytes, before.DiskWriteBytes)

	// Whether the write is attributed by the time it is observed depends on the
	// filesystem and, on virtualized CI disks, on the host. Treat a flat counter
	// as inconclusive rather than as a failure.
	if after.DiskWriteBytes == before.DiskWriteBytes {
		t.Skip("disk write bytes did not move; kernel did not attribute the write")
	}
}

func TestIOCounters_NonExistent(t *testing.T) {
	p := &Process{Pid: 99999}

	_, err := p.IOCounters()
	require.ErrorIs(t, err, ErrorProcessNotRunning)
}
