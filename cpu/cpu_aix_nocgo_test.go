// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package cpu

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// testInvoker is a mock invoker that reads fixture files from testdata/aix/.
// File naming follows FakeInvoke convention: command name + args concatenated.
type testInvoker struct {
	common.FakeInvoke
}

func (testInvoker) Command(name string, arg ...string) ([]byte, error) {
	var b strings.Builder
	b.WriteString(filepath.Base(name))
	for _, a := range arg {
		b.WriteString(a)
	}
	return os.ReadFile(filepath.Join("testdata", "aix", b.String()))
}

func (t testInvoker) CommandWithContext(_ context.Context, name string, arg ...string) ([]byte, error) {
	return t.Command(name, arg...)
}

func TestTimesWithContextAndInterval_PerCPU(t *testing.T) {
	origInvoke := invoke
	invoke = testInvoker{}
	defer func() { invoke = origInvoke }()

	// Fixture sar-u-PALL101 was captured with interval=10, count=1
	stats, err := timesWithContextAndInterval(context.Background(), true, 10)
	require.NoError(t, err)

	// Fixture has 4 per-CPU rows (cpus 0-3) plus the aggregate line (-)
	// The aggregate line has "-" as cpu field, which fails Atoi -> gets skipped
	require.Len(t, stats, 4, "expected 4 per-CPU entries")

	// CPU 0: 1 %usr, 11 %sys, 0 %wio, 88 %idle
	assert.Equal(t, "0", stats[0].CPU)
	assert.InDelta(t, 1.0, stats[0].User, 0.01)
	assert.InDelta(t, 11.0, stats[0].System, 0.01)
	assert.InDelta(t, 0.0, stats[0].Iowait, 0.01)
	assert.InDelta(t, 88.0, stats[0].Idle, 0.01)

	// CPU 1: 0 %usr, 0 %sys, 0 %wio, 100 %idle
	assert.Equal(t, "1", stats[1].CPU)
	assert.InDelta(t, 0.0, stats[1].User, 0.01)
	assert.InDelta(t, 0.0, stats[1].System, 0.01)
	assert.InDelta(t, 0.0, stats[1].Iowait, 0.01)
	assert.InDelta(t, 100.0, stats[1].Idle, 0.01)

	// CPU 2: same as CPU 1
	assert.Equal(t, "2", stats[2].CPU)
	assert.InDelta(t, 100.0, stats[2].Idle, 0.01)

	// CPU 3: same as CPU 1
	assert.Equal(t, "3", stats[3].CPU)
	assert.InDelta(t, 100.0, stats[3].Idle, 0.01)
}

func TestTimesWithContextAndInterval_Aggregate(t *testing.T) {
	origInvoke := invoke
	invoke = testInvoker{}
	defer func() { invoke = origInvoke }()

	// Fixture sar-u101 was captured with interval=10, count=1
	stats, err := timesWithContextAndInterval(context.Background(), false, 10)
	require.NoError(t, err)
	require.Len(t, stats, 1)

	assert.Equal(t, "cpu-total", stats[0].CPU)
	assert.InDelta(t, 0.0, stats[0].User, 0.01)
	assert.InDelta(t, 3.0, stats[0].System, 0.01)
	assert.InDelta(t, 0.0, stats[0].Iowait, 0.01)
	assert.InDelta(t, 96.0, stats[0].Idle, 0.01)
}
