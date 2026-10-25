// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/internal/common"
	"github.com/shirou/gopsutil/v4/net"
)

// MemoryMapsStat is not available on AIX.
type MemoryMapsStat struct{}

// MemoryInfoExStat is not available on AIX.
type MemoryInfoExStat struct{}

func readPsinfo(ctx context.Context, pid int32) (*psinfo, error) {
	f, err := os.Open(common.HostProcWithContext(ctx, strconv.Itoa(int(pid)), "psinfo"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var psi psinfo
	if err := binary.Read(f, binary.BigEndian, &psi); err != nil {
		return nil, err
	}
	return &psi, nil
}

func nullTerminatedString(b []byte) string {
	if idx := bytes.IndexByte(b, 0); idx >= 0 {
		return string(b[:idx])
	}
	return string(b)
}

func pidsWithContext(ctx context.Context) ([]int32, error) {
	dir, err := os.Open(common.HostProcWithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	pids := make([]int32, 0, len(names))
	for _, name := range names {
		pid, err := strconv.ParseInt(name, 10, 32)
		if err != nil {
			continue // skip non-numeric entries (e.g. "net", "sys")
		}
		pids = append(pids, int32(pid))
	}
	return pids, nil
}

func ProcessesWithContext(ctx context.Context) ([]*Process, error) {
	pids, err := pidsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]*Process, 0, len(pids))
	for _, pid := range pids {
		// create Process struct directly to avoid the redundant PidExists check
		ret = append(ret, &Process{Pid: pid})
	}
	return ret, nil
}

func (p *Process) PpidWithContext(ctx context.Context) (int32, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return 0, err
	}
	// psinfo stores PIDs as uint64; safe to truncate to int32 on AIX where PIDs fit in 32 bits.
	return int32(psi.Ppid), nil
}

func (p *Process) NameWithContext(ctx context.Context) (string, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return "", err
	}
	if name := nullTerminatedString(psi.Fname[:]); name != "" {
		return name, nil
	}
	// PID 0 is the swapper/idle process; its Fname is empty but ps shows "swapper".
	if p.Pid == 0 {
		return "swapper", nil
	}
	return "", nil
}

func (*Process) TgidWithContext(_ context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

func (*Process) ExeWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func (p *Process) CmdlineWithContext(ctx context.Context) (string, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return "", err
	}
	// Psargs is empty for kernel threads, fall back to Fname (the executable name),
	// which is what ps uses as well.
	if args := nullTerminatedString(psi.Psargs[:]); args != "" {
		return args, nil
	}
	if name := nullTerminatedString(psi.Fname[:]); name != "" {
		return name, nil
	}
	// PID 0 is the swapper/idle process, its Fname and Psargs are both empty
	// but ps shows "swapper".
	if p.Pid == 0 {
		return "swapper", nil
	}
	return "", nil
}

func (p *Process) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	// Psargs is empty for kernel threads, fall back to Fname (the executable name),
	// which is what ps uses as well.
	args := nullTerminatedString(psi.Psargs[:])
	if args == "" {
		args = nullTerminatedString(psi.Fname[:])
	}
	// PID 0 is the swapper/idle process, its Fname and Psargs are both empty
	// but ps shows "swapper".
	if args == "" && p.Pid == 0 {
		args = "swapper"
	}
	if args == "" {
		return nil, nil
	}
	return strings.Fields(args), nil
}

func (p *Process) createTimeWithContext(ctx context.Context) (int64, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return 0, err
	}
	return psi.Start.Sec*1000 + int64(psi.Start.Nsec)/1000000, nil
}

func (*Process) CwdWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func (p *Process) StatusWithContext(ctx context.Context) ([]string, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	return []string{convertStatusChar(string([]byte{psi.Lwp.Sname}))}, nil
}

func (*Process) ForegroundWithContext(_ context.Context) (bool, error) {
	return false, common.ErrNotImplementedError
}

func (p *Process) UidsWithContext(ctx context.Context) ([]uint32, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	// psinfo stores UIDs as uint64; safe to truncate to uint32 on AIX where UIDs fit in 32 bits.
	// real, effective, saved (psinfo doesn't expose saved, use real as fallback)
	return []uint32{uint32(psi.Uid), uint32(psi.Euid), uint32(psi.Uid)}, nil
}

func (p *Process) GidsWithContext(ctx context.Context) ([]uint32, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	// psinfo stores GIDs as uint64; safe to truncate to uint32 on AIX where GIDs fit in 32 bits.
	// real, effective, saved (psinfo doesn't expose saved, use real as fallback)
	return []uint32{uint32(psi.Gid), uint32(psi.Egid), uint32(psi.Gid)}, nil
}

func (*Process) GroupsWithContext(_ context.Context) ([]uint32, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) TerminalWithContext(_ context.Context) (string, error) {
	return "", common.ErrNotImplementedError
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return 0, err
	}
	// Returns the raw pr_nice value from lwpsinfo_t, this is not the same as what ps displays
	return int32(psi.Lwp.Nice), nil
}

func (*Process) IOniceWithContext(_ context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	return p.RlimitUsageWithContext(ctx, false)
}

func (*Process) RlimitUsageWithContext(_ context.Context, _ bool) ([]RlimitStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) NumCtxSwitchesWithContext(_ context.Context) (*NumCtxSwitchesStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) NumFDsWithContext(_ context.Context) (int32, error) {
	return 0, common.ErrNotImplementedError
}

func (p *Process) NumThreadsWithContext(ctx context.Context) (int32, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return 0, err
	}
	return int32(psi.Nlwp), nil
}

func (*Process) ThreadsWithContext(_ context.Context) (map[int32]*cpu.TimesStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	// psinfo only provides combined user+system time
	combined := float64(psi.Time.Sec) + float64(psi.Time.Nsec)/1e9
	return &cpu.TimesStat{
		CPU:    "cpu",
		User:   combined,
		System: 0,
	}, nil
}

func (*Process) CPUAffinityWithContext(_ context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {
	psi, err := readPsinfo(ctx, p.Pid)
	if err != nil {
		return nil, err
	}
	// pr_size and pr_rssize are in KB per documentation
	return &MemoryInfoStat{
		RSS: psi.Rssize * 1024,
		VMS: psi.Size * 1024,
	}, nil
}

func (*Process) MemoryInfoExWithContext(_ context.Context) (*MemoryInfoExStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) PageFaultsWithContext(_ context.Context) (*PageFaultsStat, error) {
	return nil, common.ErrNotImplementedError
}

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
	pids, err := pidsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	ret := make([]*Process, 0)
	for _, pid := range pids {
		psi, err := readPsinfo(ctx, pid)
		if err != nil {
			continue
		}
		// psinfo stores PIDs as uint64; safe to truncate to int32 on AIX where PIDs fit in 32 bits.
		if int32(psi.Ppid) == p.Pid {
			// create Process struct directly to avoid the redundant PidExists check
			ret = append(ret, &Process{Pid: pid})
		}
	}
	return ret, nil
}

func (*Process) OpenFilesWithContext(_ context.Context) ([]OpenFilesStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) ConnectionsWithContext(_ context.Context) ([]net.ConnectionStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) ConnectionsMaxWithContext(_ context.Context, _ int) ([]net.ConnectionStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) MemoryMapsWithContext(_ context.Context, _ bool) (*[]MemoryMapsStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) IOCountersWithContext(_ context.Context) (*IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func (*Process) EnvironWithContext(_ context.Context) ([]string, error) {
	return nil, common.ErrNotImplementedError
}
