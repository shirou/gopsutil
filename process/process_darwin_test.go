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
