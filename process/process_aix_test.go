// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fillPsinfo writes the Fname and Psargs as bytes in the psinfo struct and returns it
func fillPsinfo(psi psinfo, fname, psargs string) psinfo {
	copy(psi.Fname[:], fname)
	copy(psi.Psargs[:], psargs)
	return psi
}

// writeFakePsinfo serializes psi as big-endian and writes it to
// <dir>/<pid>/psinfo, creating the directory as needed.
func writeFakePsinfo(t *testing.T, dir string, psi psinfo) {
	t.Helper()
	pidDir := filepath.Join(dir, strconv.FormatUint(psi.Pid, 10))
	require.NoError(t, os.MkdirAll(pidDir, 0o755))

	var buf bytes.Buffer
	require.NoError(t, binary.Write(&buf, binary.BigEndian, &psi))
	require.NoError(t, os.WriteFile(filepath.Join(pidDir, "psinfo"), buf.Bytes(), 0o600))
}

// setupFakeProc creates a temp directory with fake psinfo files,
// sets HOST_PROC to point at it, and returns a context.
func setupFakeProc(t *testing.T, procs ...psinfo) context.Context {
	t.Helper()
	dir := t.TempDir()
	for i := range procs {
		writeFakePsinfo(t, dir, procs[i])
	}
	t.Setenv("HOST_PROC", dir)
	return context.Background()
}

// mockProc is the main test process fixture
var mockProc = fillPsinfo(psinfo{
	Pid:    1234,
	Ppid:   100,
	UID:    501,
	Euid:   502,
	Gid:    20,
	Egid:   21,
	Nlwp:   3,
	Size:   256,
	Rssize: 64,
	Start:  prTimestruc64{Sec: 1700000000, Nsec: 123000000},
	Time:   prTimestruc64{Sec: 5, Nsec: 250000000},
	Lwp:    lwpSinfo{Nice: 65, Sname: 'R'},
}, "myproc", "/usr/bin/myproc -v --flag arg")

// childProc is a child of mockProc (Ppid == mockProc.Pid)
var childProc = fillPsinfo(psinfo{
	Pid:    5678,
	Ppid:   1234,
	UID:    501,
	Euid:   501,
	Gid:    20,
	Egid:   20,
	Nlwp:   1,
	Size:   128,
	Rssize: 32,
	Lwp:    lwpSinfo{Sname: 'S'},
}, "child", "/usr/bin/child")

func TestPsinfoPpid(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	ppid, err := p.PpidWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(mockProc.Ppid), ppid)
}

func TestPsinfoName(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	name, err := p.NameWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "myproc", name)
}

func TestPsinfoCmdline(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	cmd, err := p.CmdlineWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "/usr/bin/myproc -v --flag arg", cmd)
}

func TestPsinfoCmdlineSlice(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	args, err := p.CmdlineSliceWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"/usr/bin/myproc", "-v", "--flag", "arg"}, args)
}

func TestPsinfoCmdlineSliceEmpty(t *testing.T) {
	// Both Psargs and Fname empty → nil
	ctx := setupFakeProc(t, psinfo{Pid: 9999, Ppid: 1, Lwp: lwpSinfo{Sname: 'S'}})
	p := &Process{Pid: 9999}
	args, err := p.CmdlineSliceWithContext(ctx)
	require.NoError(t, err)
	assert.Nil(t, args)
}

func TestPsinfoCmdlineFallsBackToFname(t *testing.T) {
	// Psargs empty (kernel thread) → falls back to Fname, same as ps
	ctx := setupFakeProc(t, fillPsinfo(psinfo{Pid: 9999, Ppid: 1, Lwp: lwpSinfo{Sname: 'S'}}, "kthread", ""))
	p := &Process{Pid: 9999}
	cmd, err := p.CmdlineWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "kthread", cmd)

	args, err := p.CmdlineSliceWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"kthread"}, args)
}

func TestPsinfoCreateTime(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	ct, err := p.CreateTimeWithContext(ctx)
	require.NoError(t, err)
	// 1700000000 * 1000 + 123000000 / 1000000 = 1700000000000 + 123 = 1700000000123
	assert.Equal(t, int64(1700000000123), ct)
}

func TestPsinfoStatus(t *testing.T) {
	tests := []struct {
		sname    byte
		expected string
	}{
		{'R', Running},
		{'S', Sleep},
		{'Z', Zombie},
		{'T', Stop},
		{'I', Idle},
		{'X', UnknownState}, // unknown → UnknownState
	}
	for _, tt := range tests {
		ctx := setupFakeProc(t, psinfo{Pid: 42, Ppid: 1, Lwp: lwpSinfo{Sname: tt.sname}})
		p := &Process{Pid: 42}
		status, err := p.StatusWithContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{tt.expected}, status, "sname=%c", tt.sname)
	}
}

func TestPsinfoUids(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	uids, err := p.UidsWithContext(ctx)
	require.NoError(t, err)
	// real=501, effective=502, saved=501 (fallback to real)
	assert.Equal(t, []uint32{501, 502, 501}, uids)
}

func TestPsinfoGids(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	gids, err := p.GidsWithContext(ctx)
	require.NoError(t, err)
	// real=20, effective=21, saved=20 (fallback to real)
	assert.Equal(t, []uint32{20, 21, 20}, gids)
}

func TestPsinfoNice(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	nice, err := p.NiceWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(65), nice) // raw pr_nice value
}

func TestPsinfoNumThreads(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	n, err := p.NumThreadsWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, int32(3), n)
}

func TestPsinfoTimes(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	times, err := p.TimesWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "cpu", times.CPU)
	assert.InDelta(t, 5.25, times.User, 1e-6)
	assert.InDelta(t, float64(0), times.System, 1e-6)
}

func TestPsinfoMemoryInfo(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	p := &Process{Pid: int32(mockProc.Pid)}
	mem, err := p.MemoryInfoWithContext(ctx)
	require.NoError(t, err)
	// pr_rssize=64 KB, pr_size=256 KB → bytes
	assert.Equal(t, uint64(64)*1024, mem.RSS)
	assert.Equal(t, uint64(256)*1024, mem.VMS)
}

func TestPsinfoPids(t *testing.T) {
	ctx := setupFakeProc(t, mockProc, childProc)
	pids, err := pidsWithContext(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []int32{1234, 5678}, pids)
}

func TestPsinfoPidsSkipsNonNumeric(t *testing.T) {
	dir := t.TempDir()
	writeFakePsinfo(t, dir, mockProc)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "net"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sys"), 0o755))
	t.Setenv("HOST_PROC", dir)
	pids, err := pidsWithContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []int32{1234}, pids)
}

func TestPsinfoProcessesWithContext(t *testing.T) {
	ctx := setupFakeProc(t, mockProc, childProc)
	procs, err := ProcessesWithContext(ctx)
	require.NoError(t, err)
	require.Len(t, procs, 2)

	pids := make(map[int32]struct{})
	for _, p := range procs {
		pids[p.Pid] = struct{}{}
	}
	assert.Contains(t, pids, int32(mockProc.Pid))
	assert.Contains(t, pids, int32(childProc.Pid))
}

func TestPsinfoChildren(t *testing.T) {
	ctx := setupFakeProc(t, mockProc, childProc)
	parent := &Process{Pid: int32(mockProc.Pid)}
	children, err := parent.ChildrenWithContext(ctx)
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, int32(childProc.Pid), children[0].Pid)
}

func TestPsinfoChildrenNone(t *testing.T) {
	ctx := setupFakeProc(t, mockProc)
	parent := &Process{Pid: int32(mockProc.Pid)}
	children, err := parent.ChildrenWithContext(ctx)
	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestPsinfoMissingFile(t *testing.T) {
	ctx := setupFakeProc(t) // empty proc dir
	p := &Process{Pid: 9999}
	_, err := p.NameWithContext(ctx)
	assert.Error(t, err)
}

func TestPsinfoSwapper(t *testing.T) {
	// PID 0 (swapper) has empty Fname and Psargs; should return "swapper".
	ctx := setupFakeProc(t, psinfo{Pid: 0, Lwp: lwpSinfo{Sname: 'R'}})
	p := &Process{Pid: 0}

	name, err := p.NameWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "swapper", name)

	cmd, err := p.CmdlineWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, "swapper", cmd)

	args, err := p.CmdlineSliceWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"swapper"}, args)
}
