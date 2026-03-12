// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockInvoker struct {
	outputs map[string]string
}

func (m *mockInvoker) Command(name string, arg ...string) ([]byte, error) {
	return m.CommandWithContext(context.Background(), name, arg...)
}

func (m *mockInvoker) CommandWithContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	key := name
	for _, a := range arg {
		key += " " + a
	}
	if out, ok := m.outputs[key]; ok {
		return []byte(out), nil
	}
	return nil, assert.AnError
}

// An AIX test binary would also contain common tests
// which will call inimplemented aix functions,
// so after building the test binary, you must
// use `TestProcess_AIX` prefix for running AIX tests.
func TestProcess_AIX_MemoryInfo(t *testing.T) {
	originalInvoke := invoke
	defer func() { invoke = originalInvoke }()

	mock := &mockInvoker{
		outputs: map[string]string{
			"ps -o rssize -p 1234": "RSS\n 1024\n",
			"ps -o vsz -p 1234":    "VSZ\n 2048\n",
		},
	}
	invoke = mock

	p := &Process{Pid: 1234}
	m, err := p.MemoryInfo()
	require.NoError(t, err)
	assert.Equal(t, uint64(1024*1024), m.RSS)
	assert.Equal(t, uint64(2048*1024), m.VMS)
}

func TestProcess_AIX_Times(t *testing.T) {
	originalInvoke := invoke
	defer func() { invoke = originalInvoke }()

	mock := &mockInvoker{
		outputs: map[string]string{
			"ps -o time -p 1234": "TIME\n 01:02:03\n",
		},
	}
	invoke = mock

	p := &Process{Pid: 1234}
	times, err := p.Times()
	assert.NoError(t, err)
	assert.Equal(t, &cpu.TimesStat{User: 3723.0}, times)
}

func TestProcess_AIX_CreateTime(t *testing.T) {
	originalInvoke := invoke
	defer func() { invoke = originalInvoke }()

	mock := &mockInvoker{
		outputs: map[string]string{
			"ps -o etimes -p 1234": "ELAPSED\n 3600\n",
		},
	}
	invoke = mock

	p := &Process{Pid: 1234}
	ctime, err := p.CreateTime()
	assert.NoError(t, err)
	assert.True(t, ctime > 0)
	// Check if it's roughly 1 hour ago (mock etimes is 3600)
	now := time.Now().Unix()
	expected := (now - 3600) * 1000
	assert.InDelta(t, expected, ctime, 3000)
}

func TestProcess_AIX_Cwd(t *testing.T) {
	td := t.TempDir()
	t.Setenv("HOST_PROC", td)

	p := &Process{Pid: 1234}
	pidDir := filepath.Join(td, "1234")
	err := os.MkdirAll(pidDir, 0755)
	assert.NoError(t, err)

	cwdPath := filepath.Join(pidDir, "cwd")
	target := filepath.Join(td, "target")
	err = os.MkdirAll(target, 0755)
	assert.NoError(t, err)

	err = os.Symlink(target, cwdPath)
	assert.NoError(t, err)

	cwd, err := p.Cwd()
	assert.NoError(t, err)
	assert.Equal(t, target, cwd)
}

func TestParsePsTime(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"01:02:03", 3723},
		{"02:03", 123},
		{"1-01:02:03", 86400 + 3723},
	}

	for _, tt := range tests {
		got, err := parsePsTime(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, got)
	}
}
