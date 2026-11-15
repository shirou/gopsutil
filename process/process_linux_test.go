// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package process

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func TestMemoryInfoExWithContext(t *testing.T) {
	cases := []struct {
		name      string
		statm     string
		dirty     uint64
		wantError bool
	}{
		{"normal", "100 20 3 4 0 6 0\n", 0, false},
		{"whitespace", " 100\t20  3 4 0 6 0 \n", 0, false},
		{"legacy dirty field", "100 20 3 4 0 6 7\n", 7, false},
		{"empty", "", 0, true},
		{"truncated", "100 20 3 4 0 6", 0, true},
		{"invalid data", "100 20 3 4 0 invalid 0\n", 0, true},
		{"invalid dirty", "100 20 3 4 0 6 invalid\n", 0, true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(root, "42"), 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(root, "42", "statm"), []byte(tt.statm), 0o600))
			ctx := context.WithValue(t.Context(), common.EnvKey, common.EnvMap{
				common.HostProcEnvKey: root,
			})

			p := &Process{Pid: 42}
			info, err := p.MemoryInfoExWithContext(ctx)
			if tt.wantError {
				require.Error(t, err)
				assert.Nil(t, info)
				return
			}
			require.NoError(t, err)
			pageSize := uint64(os.Getpagesize())
			assert.Equal(t, &MemoryInfoExStat{
				RSS:    20 * pageSize,
				VMS:    100 * pageSize,
				Shared: 3 * pageSize,
				Text:   4 * pageSize,
				Data:   6 * pageSize,
				Dirty:  tt.dirty * pageSize,
			}, info)
		})
	}
}

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

func TestSplitProcStat_malformed(t *testing.T) {
	// Malformed/truncated stat contents must not panic; splitProcStat
	// returns nil so callers' field-count checks can reject them.
	cases := []string{
		"",             // empty
		"123",          // no parentheses
		"123 name S 4", // no parentheses, plain fields
		"123 (name",    // missing closing paren
		"name) S 4",    // missing opening paren
		"123 (name)",   // ')' as final byte, no trailing fields
	}
	for _, content := range cases {
		t.Run(content, func(t *testing.T) {
			assert.NotPanics(t, func() {
				_ = splitProcStat([]byte(content))
			})
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

func TestForegroundWithContext_commWithSpaces(t *testing.T) {
	// comm is controlled by the process (prctl PR_SET_NAME) and may contain
	// spaces, which must not shift the pgrp/tpgid fields.
	procDir := t.TempDir()
	require.NoError(t, os.Mkdir(procDir+"/4242", 0o700))
	stat := "4242 (x 1 1 1 1) S 1 4242 4242 0 -1 4194304 0 0 0 0 0 0 0 0 20 0 1 0 100 0 0\n"
	require.NoError(t, os.WriteFile(procDir+"/4242/stat", []byte(stat), 0o600))
	t.Setenv("HOST_PROC", procDir)

	p := &Process{Pid: 4242}
	fg, err := p.ForegroundWithContext(context.Background())
	require.NoError(t, err)
	assert.False(t, fg)
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

func TestNiceFromStat(t *testing.T) {
	const (
		pid  = 4242
		nice = 7
	)

	// /proc/[pid]/stat: a normal, non real-time process reports 20+nice as its
	// scheduling priority and the niceness itself in the next field.
	fields := []string{
		"S", "1", strconv.Itoa(pid), strconv.Itoa(pid), "0", "-1", "4194560", // 3..9
		"100", "0", "0", "0", // 10..13
		"1", "2", "0", "0", // 14..17
		strconv.Itoa(20 + nice), strconv.Itoa(nice), "1", "0", "12345", // 18..22
	}
	for len(fields) < 50 { // pad up to and including field 52
		fields = append(fields, "0")
	}

	dir := t.TempDir()
	pidDir := filepath.Join(dir, strconv.Itoa(pid))
	require.NoError(t, os.MkdirAll(pidDir, 0o750))
	stat := strconv.Itoa(pid) + " (test) " + strings.Join(fields, " ") + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(stat), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pidDir, "status"), []byte("Name:\ttest\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(pidDir, "limits"), []byte(
		"Limit                     Soft Limit           Hard Limit           Units     \n"+
			"Max nice priority         0                    0                              \n"+
			"Max realtime priority     0                    0                              \n"), 0o600))
	t.Setenv("HOST_PROC", dir)

	p := &Process{Pid: pid}

	gotNice, err := p.NiceWithContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(nice), gotNice)

	rlimits, err := p.RlimitUsageWithContext(context.Background(), true)
	require.NoError(t, err)

	found := map[int32]uint64{}
	for _, rl := range rlimits {
		found[rl.Resource] = rl.Used
	}
	// RLIMIT_NICE reports the niceness in the encoding used by the limits
	// themselves, where the nice ceiling is 20-rlim_cur.
	assert.Equal(t, uint64(20-nice), found[RLIMIT_NICE])
	assert.Equal(t, uint64(0), found[RLIMIT_RTPRIO])
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
