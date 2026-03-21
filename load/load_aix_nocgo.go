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

// testInvoker is used for dependency injection in tests
var testInvoker common.Invoker

// getInvoker returns the test invoker if set, otherwise returns the default
func getInvoker() common.Invoker {
	if testInvoker != nil {
		return testInvoker
	}
	return common.Invoke{}
}

func AvgWithContext(ctx context.Context) (*AvgStat, error) {
	line, err := getInvoker().CommandWithContext(ctx, "uptime")
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
	out, err := getInvoker().CommandWithContext(ctx, "ps", "-e", "-o", "state")
	if err != nil {
		return nil, err
	}

	ret := &MiscStat{}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip header line and empty lines
		if line == "ST" || line == "STATE" || line == "S" || line == "" {
			continue
		}

		// Count processes by state (AIX process states from official docs)
		// A = Active (running or ready to run)
		// W = Swapped (not in main memory)
		// I = Intermediate (being created)
		// Z = Zombie/defunct (terminated, waiting for parent)
		// T = Stopped (trace stopped)
		// O = Nonexistent
		switch line {
		case "A", "I":
			// A = Active (running or ready to run)
			// I = Intermediate (being created via fork — transient state where
			//     the kernel is actively consuming CPU cycles, so classified as
			//     running to match AIX's own definition of active processes)
			ret.ProcsRunning++
		case "W":
			// Swapped processes (blocked/not runnable)
			ret.ProcsBlocked++
		case "Z", "T":
			// Zombie or Stopped processes - counted in total only
		}
		ret.ProcsTotal++
	}

	return ret, nil
}
