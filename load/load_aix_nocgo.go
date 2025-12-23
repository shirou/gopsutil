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

// parseVmstatLine parses a single line of vmstat output and extracts context switches, interrupts, and syscalls
// Format: r  b   avm   fre  re  pi  po  fr   sr  cy  in   sy  cs us sy id wa    pc    ec
func parseVmstatLine(line string) (ctxt, interrupts, syscalls int, err error) {
	fields := strings.Fields(line)
	if len(fields) < 13 {
		return 0, 0, 0, common.ErrNotImplementedError
	}

	// Column indices in vmstat output (0-based):
	// in = interrupts (index 10)
	// sy = system calls (index 11)
	// cs = context switches (index 12)
	if v, err := strconv.Atoi(fields[10]); err == nil {
		interrupts = v
	}
	if v, err := strconv.Atoi(fields[11]); err == nil {
		syscalls = v
	}
	if v, err := strconv.Atoi(fields[12]); err == nil {
		ctxt = v
	}

	return ctxt, interrupts, syscalls, nil
}

// SystemCallsWithContext returns the number of system calls since boot
func SystemCallsWithContext(ctx context.Context) (int, error) {
	out, err := getInvoker().CommandWithContext(ctx, "vmstat", "1", "1")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(out), "\n")
	// Last non-empty line contains the data
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		_, _, syscalls, err := parseVmstatLine(line)
		return syscalls, err
	}

	return 0, common.ErrNotImplementedError
}

// InterruptsWithContext returns the number of interrupts since boot
func InterruptsWithContext(ctx context.Context) (int, error) {
	out, err := getInvoker().CommandWithContext(ctx, "vmstat", "1", "1")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(out), "\n")
	// Last non-empty line contains the data
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		_, interrupts, _, err := parseVmstatLine(line)
		return interrupts, err
	}

	return 0, common.ErrNotImplementedError
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
		// Z = Canceled (zombie - terminated, waiting for parent)
		// T = Stopped (trace stopped)
		// O = Nonexistent
		switch line {
		case "A", "I":
			// Active or Idle processes (ready to run or awaiting startup)
			ret.ProcsRunning++
		case "W", "T", "Z":
			// Swapped, Stopped, or Zombie processes (blocked/not runnable)
			ret.ProcsBlocked++
		}
		ret.ProcsTotal++
	}

	// Get context switches from vmstat
	ctxt, _, _, err := getVmstatMetrics(ctx)
	if err == nil {
		ret.Ctxt = ctxt
	}

	return ret, nil
}

// getVmstatMetrics parses vmstat output and returns context switches, interrupts, and syscalls
func getVmstatMetrics(ctx context.Context) (int, int, int, error) {
	out, err := getInvoker().CommandWithContext(ctx, "vmstat", "1", "1")
	if err != nil {
		return 0, 0, 0, err
	}

	lines := strings.Split(string(out), "\n")
	// Last non-empty line contains the data
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		return parseVmstatLine(line)
	}

	return 0, 0, 0, common.ErrNotImplementedError
}
