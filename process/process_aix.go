// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package process

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/internal/common"
	"github.com/shirou/gopsutil/v4/net"
)

type MemoryMapsStat struct {
	Path         string `json:"path"`
	Rss          uint64 `json:"rss"`
	Size         uint64 `json:"size"`
	Pss          uint64 `json:"pss"`
	SharedClean  uint64 `json:"sharedClean"`
	SharedDirty  uint64 `json:"sharedDirty"`
	PrivateClean uint64 `json:"privateClean"`
	PrivateDirty uint64 `json:"privateDirty"`
	Referenced   uint64 `json:"referenced"`
	Anonymous    uint64 `json:"anonymous"`
	Swap         uint64 `json:"swap"`
}

type MemoryInfoExStat struct {
	RSS    uint64 `json:"rss"`
	VMS    uint64 `json:"vms"`
	Shared uint64 `json:"shared"`
	Text   uint64 `json:"text"`
	Lib    uint64 `json:"lib"`
	Data   uint64 `json:"data"`
	Dirty  uint64 `json:"dirty"`
}

func pidsWithContext(ctx context.Context) ([]int32, error) {
	panic("CRITICAL: Function not implemented")
}

func ProcessesWithContext(ctx context.Context) ([]*Process, error) {
	panic("CRITICAL: Function not implemented")
}
func (p *Process) PpidWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) NameWithContext(ctx context.Context) (string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) TgidWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ExeWithContext(ctx context.Context) (string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) CmdlineWithContext(ctx context.Context) (string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) CmdlineSliceWithContext(ctx context.Context) ([]string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) createTimeWithContext(ctx context.Context) (int64, error) {
	v, err := p.getPsField(ctx, "etimes")
	if err != nil {
		return 0, err
	}
	etimes := common.ParseUptime(v)
	return (time.Now().Unix() - int64(etimes)) * 1000, nil
}

func (p *Process) CwdWithContext(ctx context.Context) (string, error) {
	path := common.HostProcWithContext(ctx, strconv.Itoa(int(p.Pid)), "cwd")
	resolved, err := os.Readlink(path)
	if err != nil {
		// Fallback to checking if the process exists
		if _, statErr := os.Stat(common.HostProcWithContext(ctx, strconv.Itoa(int(p.Pid)))); os.IsNotExist(statErr) {
			return "", ErrorProcessNotRunning
		}
		return "", err
	}
	return resolved, nil
}

func (p *Process) StatusWithContext(ctx context.Context) ([]string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ForegroundWithContext(ctx context.Context) (bool, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) UidsWithContext(ctx context.Context) ([]uint32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) GidsWithContext(ctx context.Context) ([]uint32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) GroupsWithContext(ctx context.Context) ([]uint32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) TerminalWithContext(ctx context.Context) (string, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) NiceWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) IOniceWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) RlimitWithContext(ctx context.Context) ([]RlimitStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) RlimitUsageWithContext(ctx context.Context, gatherUsed bool) ([]RlimitStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) IOCountersWithContext(ctx context.Context) (*IOCountersStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) NumCtxSwitchesWithContext(ctx context.Context) (*NumCtxSwitchesStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) NumFDsWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) NumThreadsWithContext(ctx context.Context) (int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ThreadsWithContext(ctx context.Context) (map[int32]*cpu.TimesStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) TimesWithContext(ctx context.Context) (*cpu.TimesStat, error) {
	v, err := p.getPsField(ctx, "time")
	if err != nil {
		return nil, err
	}
	t, err := parsePsTime(v)
	if err != nil {
		return nil, err
	}
	return &cpu.TimesStat{
		User: t,
	}, nil
}

func (p *Process) CPUAffinityWithContext(ctx context.Context) ([]int32, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) MemoryInfoWithContext(ctx context.Context) (*MemoryInfoStat, error) {
	// Try rssize first (AIX specific), then fallback to rss (POSIX)
	rss, err := p.getPsField(ctx, "rssize")
	if err != nil {
		rss, err = p.getPsField(ctx, "rss")
		if err != nil {
			return nil, err
		}
	}
	vms, err := p.getPsField(ctx, "vsz")
	if err != nil {
		vms, err = p.getPsField(ctx, "vsize")
		if err != nil {
			return nil, err
		}
	}

	rssUint, _ := strconv.ParseUint(rss, 10, 64)
	vmsUint, _ := strconv.ParseUint(vms, 10, 64)

	return &MemoryInfoStat{
		RSS: rssUint * 1024,
		VMS: vmsUint * 1024,
	}, nil
}

func (p *Process) MemoryInfoExWithContext(ctx context.Context) (*MemoryInfoExStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) PageFaultsWithContext(ctx context.Context) (*PageFaultsStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ChildrenWithContext(ctx context.Context) ([]*Process, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) OpenFilesWithContext(ctx context.Context) ([]OpenFilesStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ConnectionsWithContext(ctx context.Context) ([]net.ConnectionStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) ConnectionsMaxWithContext(ctx context.Context, max int) ([]net.ConnectionStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) MemoryMapsWithContext(ctx context.Context, grouped bool) (*[]MemoryMapsStat, error) {
	panic("CRITICAL: Function not implemented")
}

func (p *Process) EnvironWithContext(ctx context.Context) ([]string, error) {
	panic("CRITICAL: Function not implemented")
}

// Internal functions

func (p *Process) getPsField(ctx context.Context, field string) (string, error) {
	out, err := invoke.CommandWithContext(ctx, "ps", "-o", field, "-p", strconv.Itoa(int(p.Pid)))
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected ps output: %s", out)
	}
	return strings.TrimSpace(lines[1]), nil
}

func parsePsTime(s string) (float64, error) {
	var days, hours, mins, secs float64
	var err error

	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		days, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		s = parts[1]
	}

	parts := strings.Split(s, ":")
	if len(parts) == 3 {
		hours, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		mins, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
		secs, err = strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
	} else if len(parts) == 2 {
		mins, err = strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, err
		}
		secs, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	return days*86400 + hours*3600 + mins*60 + secs, nil
}
