// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package process

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillFromfdWithContext(t *testing.T) {
	type expect struct {
		numFDs    int32
		openFiles []OpenFilesStat
		err       error
	}
	type testCase struct {
		name     string
		pid      int32
		expected *expect
	}

	cases := []testCase{
		{
			name: "good path",
			pid:  1,
			expected: &expect{
				numFDs: 3,
				openFiles: []OpenFilesStat{
					{
						Path: "/foo",
						Fd:   0,
					},
					{
						Path: "/bar",
						Fd:   1,
					},
					{
						Path: "/baz",
						Fd:   2,
					},
				},
			},
		},
	}

	t.Setenv("HOST_PROC", "testdata/linux")
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProcess(tt.pid)
			require.NoError(t, err)
			numFDs, openFiles, err := p.fillFromfdWithContext(context.TODO())
			if tt.expected.err != nil {
				assert.ErrorContains(t, err, tt.expected.err.Error())
				return
			}

			//nolint:testifylint // false positive
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.numFDs, numFDs)
			assert.ElementsMatch(t, tt.expected.openFiles, openFiles)
		})
	}
}

func TestSplitProcStat(t *testing.T) {
	expectedFieldsNum := 53
	statLineContent := make([]string, expectedFieldsNum-1)
	for i := 0; i < expectedFieldsNum-1; i++ {
		statLineContent[i] = strconv.Itoa(i + 1)
	}

	cases := []string{
		"ok",
		"ok)",
		"(ok",
		"ok )",
		"ok )(",
		"ok )()",
		"() ok )()",
		"() ok (()",
		" ) ok )",
		"(ok) (ok)",
	}

	consideredFields := []int{4, 7, 10, 11, 12, 13, 14, 15, 18, 22, 42}

	commandNameIndex := 2
	for _, expectedName := range cases {
		statLineContent[commandNameIndex-1] = "(" + expectedName + ")"
		statLine := strings.Join(statLineContent, " ")
		t.Run("name: "+expectedName, func(t *testing.T) {
			parsedStatLine := splitProcStat([]byte(statLine))
			assert.Equal(t, expectedName, parsedStatLine[commandNameIndex])
			for _, idx := range consideredFields {
				expected := strconv.Itoa(idx)
				parsed := parsedStatLine[idx]
				assert.Equalf(
					t, expected, parsed,
					"field %d (index from 1 as in man proc) must be %q but %q is received",
					idx, expected, parsed,
				)
			}
		})
	}
}

func TestSplitProcStat_fromFile(t *testing.T) {
	pids, err := os.ReadDir("testdata/linux/")
	require.NoError(t, err)
	t.Setenv("HOST_PROC", "testdata/linux")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		statFile := fmt.Sprintf("testdata/linux/%d/stat", pid)
		if _, err := os.Stat(statFile); err != nil {
			continue
		}
		contents, err := os.ReadFile(statFile)
		require.NoError(t, err)

		pidStr := strconv.Itoa(int(pid))

		ppid := "68044" // TODO: how to pass ppid to test?

		fields := splitProcStat(contents)
		assert.Equal(t, pidStr, fields[1])
		assert.Equal(t, "test(cmd).sh", fields[2])
		assert.Equal(t, "S", fields[3])
		assert.Equal(t, ppid, fields[4])
		assert.Equal(t, pidStr, fields[5]) // pgrp
		assert.Equal(t, ppid, fields[6])   // session
		assert.Equal(t, pidStr, fields[8]) // tpgrp
		assert.Equal(t, "20", fields[18])  // priority
		assert.Equal(t, "1", fields[20])   // num threads
		assert.Equal(t, "0", fields[52])   // exit code
	}
}

func TestFillFromCommWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/linux/")
	require.NoError(t, err)
	t.Setenv("HOST_PROC", "testdata/linux")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("testdata/linux/%d/status", pid)); err != nil {
			continue
		}
		p, _ := NewProcess(int32(pid))
		assert.NoError(t, p.fillFromCommWithContext(context.Background()))
	}
}

func TestFillFromStatusWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/linux/")
	require.NoError(t, err)
	t.Setenv("HOST_PROC", "testdata/linux")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("testdata/linux/%d/status", pid)); err != nil {
			continue
		}
		p, _ := NewProcess(int32(pid))
		assert.NoError(t, p.fillFromStatus())
	}
}

func Benchmark_fillFromCommWithContext(b *testing.B) {
	b.Setenv("HOST_PROC", "testdata/linux")
	pid := 1060
	p, _ := NewProcess(int32(pid))
	for i := 0; i < b.N; i++ {
		p.fillFromCommWithContext(context.Background())
	}
}

func Benchmark_fillFromStatusWithContext(b *testing.B) {
	b.Setenv("HOST_PROC", "testdata/linux")
	pid := 1060
	p, _ := NewProcess(int32(pid))
	for i := 0; i < b.N; i++ {
		p.fillFromStatus()
	}
}

func TestFillFromTIDStatWithContext_lx_brandz(t *testing.T) {
	pids, err := os.ReadDir("testdata/lx_brandz/")
	require.NoError(t, err)
	t.Setenv("HOST_PROC", "testdata/lx_brandz")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("testdata/lx_brandz/%d/stat", pid)); err != nil {
			continue
		}
		p, _ := NewProcess(int32(pid))
		_, _, cpuTimes, _, _, _, _, err := p.fillFromTIDStat(-1)
		require.NoError(t, err)
		assert.Zero(t, cpuTimes.Iowait)
	}
}

func TestProcessMemoryMaps(t *testing.T) {
	t.Setenv("HOST_PROC", "testdata/linux")
	pid := 1
	p, err := NewProcess(int32(pid))
	require.NoError(t, err)
	maps, err := p.MemoryMaps(false)
	require.NoError(t, err)

	expected := &[]MemoryMapsStat{
		{
			"[vvar]",
			0,
			1,
			0,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
		},
		{
			"",
			0,
			1,
			2,
			3,
			4,
			0,
			6,
			7,
			8,
			9,
		},
		{
			"[vdso]",
			0,
			1,
			2,
			3,
			4,
			5,
			0,
			7,
			8,
			9,
		},
		{
			"/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1",
			0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			0,
			9,
		},
	}

	require.Equal(t, expected, maps)
}
