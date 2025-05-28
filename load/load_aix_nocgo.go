// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package load

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var separator = regexp.MustCompile(`,?\s+`)

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	line, err := invoke.CommandWithContext(ctx, "uptime")
	if err != nil {
		return nil, err
	}

	idx := bytes.Index(line, []byte("load average:"))
	if idx < 0 {
		return nil, common.ErrNotImplementedError
	}
	ret := &AvgStat{}

	p := separator.Split(string(line[idx:]), 5)
	if 4 < len(p) && p[0] == "load" && p[1] == "average:" {
		if t, err := strconv.ParseFloat(p[2], 64); err == nil {
			ret.Load1 = t
		}
		if t, err := strconv.ParseFloat(p[3], 64); err == nil {
			ret.Load5 = t
		}
		if t, err := strconv.ParseFloat(p[4], 64); err == nil {
			ret.Load15 = t
		}
		return ret, nil
	}

	return nil, common.ErrNotImplementedError
}

func MiscWithContext(ctx context.Context) (*MiscStat, error) {
	out, err := invoke.CommandWithContext(ctx, "ps", "-Ao", "state")
	if err != nil {
		return nil, err
	}

	ret := &MiscStat{}
	for _, line := range strings.Split(string(out), "\n") {
		ret.ProcsTotal++
		switch line {
		case "R":
		case "A":
			ret.ProcsRunning++
		case "T":
			ret.ProcsBlocked++
		default:
			continue
		}
	}
	return ret, nil
}
