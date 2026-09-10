// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package cpu

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func exTimesContext(t *testing.T, data string) context.Context {
	t.Helper()
	proc := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(proc, "stat"), []byte(data), 0o600))
	return context.WithValue(t.Context(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: proc})
}

func TestExLinuxTimes(t *testing.T) {
	ctx := exTimesContext(t, "cpu  101 102 103 104 105 106 107 108 109 110\n"+
		"cpu0 1 2 3 4 5 6 7 8 9 10\ncpu7 11 12 13 14 15 16 17 18 19 20\n"+
		"intr 999\ncpu9 21 22 23 24 25 26 27 28 29 30\n")
	t.Setenv("HOST_PROC", t.TempDir())
	ex := NewExLinux()
	total, err := ex.TimesWithContext(ctx, false)
	require.NoError(t, err)
	require.Equal(t, []TimesStatEx{{
		CPU: "cpu-total", User: 101, Nice: 102, System: 103, Idle: 104,
		Iowait: 105, Irq: 106, Softirq: 107, Steal: 108, Guest: 109, GuestNice: 110,
	}}, total)
	cpus, err := ex.TimesWithContext(ctx, true)
	require.NoError(t, err)
	require.Equal(t, []TimesStatEx{
		{CPU: "cpu0", User: 1, Nice: 2, System: 3, Idle: 4, Iowait: 5, Irq: 6, Softirq: 7, Steal: 8, Guest: 9, GuestNice: 10},
		{CPU: "cpu7", User: 11, Nice: 12, System: 13, Idle: 14, Iowait: 15, Irq: 16, Softirq: 17, Steal: 18, Guest: 19, GuestNice: 20},
	}, cpus)

	t.Setenv("HOST_PROC", ctx.Value(common.EnvKey).(common.EnvMap)[common.HostProcEnvKey])
	background, err := ex.Times(false)
	require.NoError(t, err)
	assert.Equal(t, total, background)
}

func TestExLinuxTimesPrecision(t *testing.T) {
	clocks := ClocksPerSec
	ClocksPerSec = 37
	t.Cleanup(func() { ClocksPerSec = clocks })
	var previous uint64
	for i, value := range []uint64{1 << 53, 1<<53 + 1, math.MaxUint64} {
		ctx := exTimesContext(t, fmt.Sprintf("cpu %d 2 3 4 5 6 7 8 9 %d\n", value, uint64(math.MaxUint64)))
		times, err := NewExLinux().TimesWithContext(ctx, false)
		require.NoError(t, err)
		require.Len(t, times, 1)
		assert.Equal(t, value, times[0].User)
		assert.Equal(t, uint64(math.MaxUint64), times[0].GuestNice)
		if i == 1 {
			assert.Equal(t, uint64(1), times[0].User-previous)
		}
		previous = times[0].User
	}
}

func TestExLinuxTimesColumns(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counters string
		want     TimesStatEx
	}{
		{"four", "1 2 3 4", TimesStatEx{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4}},
		{"seven", "1 2 3 4 5 6 7", TimesStatEx{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4, Iowait: 5, Irq: 6, Softirq: 7}},
		{"steal", "1 2 3 4 5 6 7 8", TimesStatEx{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4, Iowait: 5, Irq: 6, Softirq: 7, Steal: 8}},
		{"guest", "1 2 3 4 5 6 7 8 9", TimesStatEx{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4, Iowait: 5, Irq: 6, Softirq: 7, Steal: 8, Guest: 9}},
		{"extra", "1 2 3 4 5 6 7 8 9 10 11 12", TimesStatEx{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4, Iowait: 5, Irq: 6, Softirq: 7, Steal: 8, Guest: 9, GuestNice: 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := exTimesContext(t, "cpu\t"+tc.counters+"\n")
			stats, err := NewExLinux().TimesWithContext(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, []TimesStatEx{tc.want}, stats)
		})
	}
}

func TestExLinuxTimesErrors(t *testing.T) {
	for _, tc := range []struct {
		name, line string
		cause      error
	}{
		{"short", "cpu 1 2 3", nil},
		{"not_cpu", "btime 1 2 3 4", nil},
		{"negative", "cpu -1 2 3 4", strconv.ErrSyntax},
		{"fraction", "cpu 1.5 2 3 4", strconv.ErrSyntax},
		{"overflow", "cpu 18446744073709551616 2 3 4", strconv.ErrRange},
		{"optional", "cpu 1 2 3 4 5 6 7 8 9 invalid", strconv.ErrSyntax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stats, err := NewExLinux().TimesWithContext(exTimesContext(t, tc.line+"\n"), false)
			require.Error(t, err)
			assert.Nil(t, stats)
			if tc.cause != nil {
				require.ErrorIs(t, err, tc.cause)
			}
		})
	}
	ctx := exTimesContext(t, "cpu 1 2 3 4 5 6 7\ncpu0 1 2 3 4\ncpu1 invalid 2 3 4\n")
	stats, err := NewExLinux().TimesWithContext(ctx, true)
	require.ErrorIs(t, err, strconv.ErrSyntax)
	assert.Nil(t, stats)
}

func TestExLinuxTimesMissingAndEmpty(t *testing.T) {
	missing := context.WithValue(t.Context(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: t.TempDir()})
	empty := exTimesContext(t, "")
	badProc := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(badProc, "stat"), 0o700))
	unreadable := context.WithValue(t.Context(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: badProc})
	for _, percpu := range []bool{false, true} {
		stats, err := NewExLinux().TimesWithContext(missing, percpu)
		require.ErrorIs(t, err, os.ErrNotExist)
		assert.Nil(t, stats)
		stats, err = NewExLinux().TimesWithContext(unreadable, percpu)
		require.Error(t, err)
		assert.Nil(t, stats)
		legacy, err := TimesWithContext(missing, percpu)
		require.NoError(t, err)
		assert.Equal(t, []TimesStat{}, legacy)

		stats, err = NewExLinux().TimesWithContext(empty, percpu)
		require.NoError(t, err)
		assert.Equal(t, []TimesStatEx{}, stats)
	}
}

func TestTimesStatExString(t *testing.T) {
	want := TimesStatEx{CPU: `cpu"0`, User: math.MaxUint64, GuestNice: 1<<53 + 1}
	data := want.String()
	var got TimesStatEx
	require.NoError(t, json.Unmarshal([]byte(data), &got))
	assert.Equal(t, want, got)
	assert.Contains(t, data, `"user":18446744073709551615`)
}

func TestExLinuxTimesLineBoundaries(t *testing.T) {
	for _, suffix := range []string{"", "\n", "\nintr " + strings.Repeat("0 ", 65536) + "\n"} {
		ctx := exTimesContext(t, "cpu 1 2 3 4\ncpu0 11 12 13 14"+suffix)
		times, err := NewExLinux().TimesWithContext(ctx, true)
		require.NoError(t, err)
		assert.Equal(t, []TimesStatEx{{CPU: "cpu0", User: 11, Nice: 12, System: 13, Idle: 14}}, times)
	}
	ctx := exTimesContext(t, "cpu 1 2 3 4")
	stats, err := NewExLinux().TimesWithContext(ctx, false)
	require.NoError(t, err)
	assert.Equal(t, []TimesStatEx{{CPU: "cpu-total", User: 1, Nice: 2, System: 3, Idle: 4}}, stats)
	stats, err = NewExLinux().TimesWithContext(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, []TimesStatEx{}, stats)
}
