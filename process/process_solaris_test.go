// SPDX-License-Identifier: BSD-3-Clause
//go:build solaris

package process

import (
	"context"
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

func TestProcess(t *testing.T) {
	t.Run("TestProcess_Solaris_MemoryInfo", func(t *testing.T) {
		originalInvoke := invoke
		defer func() { invoke = originalInvoke }()

		mock := &mockInvoker{
			outputs: map[string]string{
				"ps -o rss -p 1234": "RSS\n 1024\n",
				"ps -o vsz -p 1234": "VSZ\n 2048\n",
			},
		}
		invoke = mock

		p := &Process{Pid: 1234}
		m, err := p.MemoryInfoWithContext(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(1024*1024), m.RSS)
		assert.Equal(t, uint64(2048*1024), m.VMS)
	})

	t.Run("TestProcess_Solaris_Times", func(t *testing.T) {
		originalInvoke := invoke
		defer func() { invoke = originalInvoke }()

		mock := &mockInvoker{
			outputs: map[string]string{
				"ps -o time -p 1234": "TIME\n 01:02:03\n",
			},
		}
		invoke = mock

		p := &Process{Pid: 1234}
		times, err := p.TimesWithContext(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, &cpu.TimesStat{User: 3723.0}, times)
	})

	t.Run("TestProcess_Solaris_CreateTime", func(t *testing.T) {
		originalInvoke := invoke
		defer func() { invoke = originalInvoke }()

		mock := &mockInvoker{
			outputs: map[string]string{
				"ps -o etime -p 1234": "ELAPSED\n 01:00:00\n",
			},
		}
		invoke = mock

		p := &Process{Pid: 1234}
		ctime, err := p.createTimeWithContext(context.Background())
		assert.NoError(t, err)
		assert.True(t, ctime > 0)
		// mock etime is 1 hour (3600 seconds)
		now := time.Now().Unix()
		expected := (now - 3600) * 1000
		assert.InDelta(t, expected, ctime, 3000)
	})
}
