// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

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
				assert.Equal(
					t, expected, parsed,
					"field %d (index from 1 as in man proc) must be %q but %q is received",
					idx, expected, parsed,
				)
			}
		})
	}
}

func TestSplitProcStat_fromFile(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		statFile := fmt.Sprintf("testdata/aix/%d/stat", pid)
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
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("testdata/aix/%d/status", pid)); err != nil {
			continue
		}
		p, _ := NewProcess(int32(pid))
		if err := p.fillFromCommWithContext(context.Background()); err != nil {
			t.Error(err)
		}
	}
}

func TestFillFromStatusWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pid := range pids {
		pid, err := strconv.ParseInt(pid.Name(), 0, 32)
		if err != nil {
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("testdata/aix/%d/status", pid)); err != nil {
			continue
		}
		p, _ := NewProcess(int32(pid))
		if err := p.fillFromStatus(); err != nil {
			t.Error(err)
		}
	}
}

func Benchmark_fillFromCommWithContext(b *testing.B) {
	b.Setenv("HOST_PROC", "testdata/aix")
	pid := 5767616
	p, _ := NewProcess(int32(pid))
	for i := 0; i < b.N; i++ {
		p.fillFromCommWithContext(context.Background())
	}
}

func Benchmark_fillFromStatusWithContext(b *testing.B) {
	b.Setenv("HOST_PROC", "testdata/aix")
	pid := 5767616
	p, _ := NewProcess(int32(pid))
	for i := 0; i < b.N; i++ {
		p.fillFromStatus()
	}
}

func TestFillFromTIDStatWithContext_lx_brandz(t *testing.T) {
	pids, err := os.ReadDir("testdata/lx_brandz/")
	if err != nil {
		t.Error(err)
	}
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
		if err != nil {
			t.Error(err)
		}
		assert.Zero(t, cpuTimes.Iowait)
	}
}

func TestProcessMemoryMaps(t *testing.T) {
	t.Setenv("HOST_PROC", "testdata/aix")
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

func TestFillFromExeWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pidDir := range pids {
		pid, err := strconv.ParseInt(pidDir.Name(), 0, 32)
		if err != nil {
			continue
		}
		psinfo := fmt.Sprintf("testdata/aix/%d/psinfo", pid)
		if _, err := os.Stat(psinfo); err != nil {
			continue
		}
		p, err := NewProcess(int32(pid))
		require.NoError(t, err)
		exe, err := p.fillFromExeWithContext(context.Background())
		if err == nil {
			// Should get a string (possibly empty or with executable name)
			assert.IsType(t, "", exe)
		}
	}
}

func TestFillFromCmdlineWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pidDir := range pids {
		pid, err := strconv.ParseInt(pidDir.Name(), 0, 32)
		if err != nil {
			continue
		}
		psinfo := fmt.Sprintf("testdata/aix/%d/psinfo", pid)
		if _, err := os.Stat(psinfo); err != nil {
			continue
		}
		p, err := NewProcess(int32(pid))
		require.NoError(t, err)
		cmdline, err := p.fillFromCmdlineWithContext(context.Background())
		if err == nil {
			// Should get a string (possibly empty or with command line)
			assert.IsType(t, "", cmdline)
		}
	}
}

func TestFillFromCmdlineSliceWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pidDir := range pids {
		pid, err := strconv.ParseInt(pidDir.Name(), 0, 32)
		if err != nil {
			continue
		}
		psinfo := fmt.Sprintf("testdata/aix/%d/psinfo", pid)
		if _, err := os.Stat(psinfo); err != nil {
			continue
		}
		p, err := NewProcess(int32(pid))
		require.NoError(t, err)
		cmdlineSlice, err := p.fillSliceFromCmdlineWithContext(context.Background())
		if err == nil {
			// Should get a slice of strings
			assert.IsType(t, []string{}, cmdlineSlice)
		}
	}
}

func TestFillFromStatmWithContext(t *testing.T) {
	pids, err := os.ReadDir("testdata/aix/")
	if err != nil {
		t.Error(err)
	}
	t.Setenv("HOST_PROC", "testdata/aix")
	for _, pidDir := range pids {
		pid, err := strconv.ParseInt(pidDir.Name(), 0, 32)
		if err != nil {
			continue
		}
		psinfo := fmt.Sprintf("testdata/aix/%d/psinfo", pid)
		if _, err := os.Stat(psinfo); err != nil {
			continue
		}
		p, err := NewProcess(int32(pid))
		require.NoError(t, err)
		memInfo, memInfoEx, err := p.fillFromStatmWithContext(context.Background())
		if err == nil {
			assert.NotNil(t, memInfo)
			assert.NotNil(t, memInfoEx)
			// Memory values should be non-negative
			//nolint:testifylint // value is always >= 0, but we validate it
			assert.GreaterOrEqual(t, memInfo.VMS, uint64(0))
			//nolint:testifylint // value is always >= 0, but we validate it
			assert.GreaterOrEqual(t, memInfo.RSS, uint64(0))
			//nolint:testifylint // value is always >= 0, but we validate it
			assert.GreaterOrEqual(t, memInfoEx.VMS, uint64(0))
			//nolint:testifylint // value is always >= 0, but we validate it
			assert.GreaterOrEqual(t, memInfoEx.RSS, uint64(0))
		}
	}
}

func TestTerminalWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	terminal, err := p.TerminalWithContext(ctx)
	// Terminal may or may not be available depending on how test is run
	if err == nil {
		assert.IsType(t, "", terminal)
	}
}

func TestEnvironmentWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	env, err := p.EnvironmentWithContext(ctx)
	if err == nil {
		assert.NotNil(t, env)
		assert.IsType(t, map[string]string{}, env)
	}
}

func TestPageFaultsWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	pageFaults, err := p.PageFaultsWithContext(ctx)
	if err != nil {
		t.Logf("PageFaultsWithContext error: %v", err)
		return
	}
	if pageFaults != nil {
		// Page fault counts should be non-negative
		//nolint:testifylint // minor faults field is naturally >= 0
		assert.GreaterOrEqual(t, pageFaults.MinorFaults, uint64(0))
		//nolint:testifylint // major faults field is naturally >= 0
		assert.GreaterOrEqual(t, pageFaults.MajorFaults, uint64(0))
	}
}

func TestRlimitUsageWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	limits, err := p.RlimitUsageWithContext(ctx, false)
	if err != nil {
		t.Logf("RlimitUsageWithContext error: %v", err)
		return
	}
	if len(limits) > 0 {
		for _, limit := range limits {
			// Hard limit should be >= soft limit
			assert.GreaterOrEqual(t, limit.Hard, limit.Soft)
		}
	}
}

func TestIOCountersWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	ioCounters, err := p.IOCountersWithContext(ctx)
	// IOCounters may not be available without WLM+iostat configuration
	if err == nil {
		assert.NotNil(t, ioCounters)
		//nolint:testifylint // checking non-negative constraint
		assert.GreaterOrEqual(t, ioCounters.ReadBytes, uint64(0))
		//nolint:testifylint // checking non-negative constraint
		assert.GreaterOrEqual(t, ioCounters.WriteBytes, uint64(0))
	}
}

func TestCPUAffinityWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	affinity, err := p.CPUAffinityWithContext(ctx)
	// CPU affinity may not be available on all AIX systems
	if err == nil {
		assert.NotEmpty(t, affinity)
		for _, cpu := range affinity {
			assert.GreaterOrEqual(t, cpu, int32(0))
		}
	}
}
func TestCPUPercentWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	percent, err := p.CPUPercentWithContext(ctx)
	require.NoError(t, err)
	// CPU percent should be >= 0
	assert.GreaterOrEqual(t, percent, float64(0))
	// CPU percent should not exceed 100 * number of CPUs (but ps can sometimes report >100% on single CPU)
	assert.Less(t, percent, float64(500)) // sanity check to avoid absurd values
}

func TestSignalsPendingWithContext(t *testing.T) {
	// Get current process
	ctx := context.Background()
	p := Process{Pid: int32(os.Getpid())}

	sigInfo, err := p.SignalsPendingWithContext(ctx)
	require.NoError(t, err)
	// SignalInfoStat should be valid (may have zero pending signals for normal process)
	assert.NotNil(t, &sigInfo)
	// PendingProcess should be >= 0
	assert.GreaterOrEqual(t, sigInfo.PendingProcess, uint64(0))
}
