// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package load

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// mockInvoker returns canned output for specific commands.
type mockInvoker struct {
	common.Invoke
	output map[string][]byte
}

func (m mockInvoker) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	var b strings.Builder
	b.WriteString(name)
	for _, a := range arg {
		b.WriteString(" ")
		b.WriteString(a)
	}
	key := b.String()
	if out, ok := m.output[key]; ok {
		return out, nil
	}
	return nil, common.ErrNotImplementedError
}

func TestMiscWithContext_ProcessStates(t *testing.T) {
	// Mock ps -e -o state output with known process states:
	//   3 Active (A), 1 Idle (I) -> ProcsRunning = 4
	//   2 Swapped (W) -> ProcsBlocked = 2
	//   1 Zombie (Z), 1 Stopped (T) -> counted in total only
	//   Total = 8
	psOutput := "ST\nA\nA\nI\nW\nZ\nT\nA\nW\n"

	origInvoke := invoke
	invoke = mockInvoker{
		output: map[string][]byte{
			"ps -e -o state": []byte(psOutput),
		},
	}
	defer func() { invoke = origInvoke }()

	misc, err := MiscWithContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 4, misc.ProcsRunning, "3 A + 1 I = 4 running")
	assert.Equal(t, 2, misc.ProcsBlocked, "2 W = 2 blocked")
	assert.Equal(t, 8, misc.ProcsTotal, "8 total processes")
}
