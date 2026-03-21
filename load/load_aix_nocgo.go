// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package load

import (
	"bytes"
	"context"
	"errors"
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
		// I = Idle (waiting for startup)
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

	// Get context switches, interrupts, and syscalls from vmstat.
	// These are cumulative since-boot counters from vmstat -s.
	ctxt, interrupts, syscalls, err := getVmstatMetrics(ctx)
	if err == nil {
		ret.Ctxt = ctxt
		ret.Interrupts = interrupts
		ret.SysCalls = syscalls
	}

	return ret, nil
}

// getVmstatMetrics runs vmstat -s and returns cumulative since-boot counters
// for context switches, device interrupts, and syscalls.
func getVmstatMetrics(ctx context.Context) (ctxt, interrupts, syscalls int, err error) {
	out, err := getInvoker().CommandWithContext(ctx, "vmstat", "-s")
	if err != nil {
		return 0, 0, 0, err
	}

	// vmstat -s output format: <whitespace><number> <description>
	// Example lines:
	//   5842393706 cpu context switches
	//      33412179 device interrupts
	//  12918944607 syscalls
	lines := strings.Split(string(out), "\n")
	foundCtxt, foundInterrupts, foundSyscalls := false, false, false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split into number and description at the first space after the number
		idx := strings.IndexByte(line, ' ')
		if idx < 0 {
			continue
		}
		valStr := line[:idx]
		desc := strings.TrimSpace(line[idx+1:])

		v, parseErr := strconv.Atoi(valStr)
		if parseErr != nil {
			continue
		}

		switch desc {
		case "cpu context switches":
			ctxt = v
			foundCtxt = true
		case "device interrupts":
			interrupts = v
			foundInterrupts = true
		case "syscalls":
			syscalls = v
			foundSyscalls = true
		}
	}

	if !foundCtxt || !foundInterrupts || !foundSyscalls {
		return ctxt, interrupts, syscalls, errors.New("incomplete vmstat output: missing expected counters")
	}
	return ctxt, interrupts, syscalls, nil
}
