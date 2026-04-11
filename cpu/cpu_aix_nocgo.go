// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package cpu

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// TimesWithContext returns CPU time statistics using a default 1-second sar
// sample window. On AIX nocgo, sar returns instantaneous percentage values
// (usr/sys/wio/idle summing to 100%), not cumulative tick counters like other
// platforms. See timesWithContextAndInterval for details.
func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	return timesWithContextAndInterval(ctx, percpu, 1)
}

// timesWithContextAndInterval runs sar with the specified sample interval in
// seconds and parses the output into TimesStat values. On AIX nocgo, sar
// returns CPU usage as percentages over the sample window, not cumulative
// counters. This means:
//   - The returned values are bounded 0-100 and sum to ~100%
//   - Delta math (subtracting two snapshots) does NOT apply
//   - The interval controls the sar measurement window, not a sleep between snapshots
//
// This is called by TimesWithContext (with a 1-second default) and by
// aixPercent (with the caller's requested interval) to honor the interval
// parameter in PercentWithContext.
func timesWithContextAndInterval(ctx context.Context, percpu bool, intervalSeconds int) ([]TimesStat, error) {
	if intervalSeconds < 1 {
		intervalSeconds = 1
	}
	sarInterval := strconv.Itoa(intervalSeconds)

	var ret []TimesStat
	if percpu {
		perOut, err := invoke.CommandWithContext(ctx, "sar", "-u", "-P", "ALL", sarInterval, "1")
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(perOut), "\n")
		if len(lines) < 6 {
			return []TimesStat{}, common.ErrNotImplementedError
		}

		hp := strings.Fields(lines[5]) // headers
		for l := 6; l < len(lines)-1; l++ {
			v := strings.Fields(lines[l]) // values
			if len(v) == 0 {
				continue
			}

			// Determine the CPU field position: first line has a timestamp
			// prefix, continuation lines do not
			cpuField := strings.TrimSpace(v[0])
			if l == 6 && len(v) > 1 {
				cpuField = strings.TrimSpace(v[1])
			}
			if _, err := strconv.Atoi(cpuField); err != nil {
				continue
			}

			ct := &TimesStat{}
			for i, header := range hp {
				if i >= len(v) {
					break
				}

				// Position variable for v
				pos := i
				// There is a missing field at the beginning of all but the first line
				// so adjust the position
				if l > 6 {
					pos = i - 1
				}
				// We don't want invalid positions
				if pos < 0 {
					continue
				}

				if t, err := strconv.ParseFloat(v[pos], 64); err == nil {
					switch header {
					case `cpu`:
						ct.CPU = strconv.FormatFloat(t, 'f', -1, 64)
					case `%usr`:
						ct.User = t
					case `%sys`:
						ct.System = t
					case `%wio`:
						ct.Iowait = t
					case `%idle`:
						ct.Idle = t
					}
				}
			}
			ret = append(ret, *ct)
		}
	} else {
		out, err := invoke.CommandWithContext(ctx, "sar", "-u", sarInterval, "1")
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) < 5 {
			return []TimesStat{}, common.ErrNotImplementedError
		}

		ct := &TimesStat{CPU: "cpu-total"}
		h := strings.Fields(lines[len(lines)-3]) // headers
		v := strings.Fields(lines[len(lines)-2]) // values
		for i, header := range h {
			if t, err := strconv.ParseFloat(v[i], 64); err == nil {
				switch header {
				case `%usr`:
					ct.User = t
				case `%sys`:
					ct.System = t
				case `%wio`:
					ct.Iowait = t
				case `%idle`:
					ct.Idle = t
				}
			}
		}

		ret = append(ret, *ct)
	}

	return ret, nil
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	out, err := invoke.CommandWithContext(ctx, "prtconf")
	if err != nil {
		return nil, err
	}

	ret := InfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "Number Of Processors:"):
			p := strings.Fields(line)
			if len(p) > 3 {
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					ret.Cores = int32(t)
				}
			}
		case strings.HasPrefix(line, "Processor Clock Speed:"):
			p := strings.Fields(line)
			if len(p) > 4 {
				if t, err := strconv.ParseFloat(p[3], 64); err == nil {
					switch strings.ToUpper(p[4]) {
					case "MHZ":
						ret.Mhz = t
					case "GHZ":
						ret.Mhz = t * 1000.0
					case "KHZ":
						ret.Mhz = t / 1000.0
					default:
						ret.Mhz = t
					}
				}
			}
		case strings.HasPrefix(line, "System Model:"):
			p := strings.Split(string(line), ":")
			if p != nil {
				ret.VendorID = strings.TrimSpace(p[1])
			}
		case strings.HasPrefix(line, "Processor Type:"):
			p := strings.Split(string(line), ":")
			if p != nil {
				c := strings.Split(string(p[1]), "_")
				if c != nil {
					ret.Family = strings.TrimSpace(c[0])
					ret.Model = strings.TrimSpace(c[1])
				}
			}
		}
	}
	return []InfoStat{ret}, nil
}

func CountsWithContext(ctx context.Context, _ bool) (int, error) {
	info, err := InfoWithContext(ctx)
	if err == nil {
		return int(info[0].Cores), nil
	}
	return 0, err
}

// aixPercent returns CPU busy percentages by calling sar with the requested
// interval. On AIX nocgo, sar returns instantaneous percentage values
// (usr/sys/wio/idle summing to 100%), so delta math is not needed — the
// percentage is computed directly as User + System, excluding Iowait, to
// match the getAllBusy semantics used on other platforms.
//
// The interval parameter is converted to integer seconds and passed to sar
// as the measurement window (minimum 1 second). This honors the interval
// parameter from PercentWithContext rather than always using a fixed window.
//
// This function is only compiled for nocgo builds. The CGO build returns
// ErrNotImplementedError so the caller falls through to the standard
// delta-based calculation (since the CGO path returns cumulative tick
// counters from perfstat, not percentages).
func aixPercent(ctx context.Context, interval time.Duration, percpu bool) ([]float64, error) {
	seconds := int(interval.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	times, err := timesWithContextAndInterval(ctx, percpu, seconds)
	if err != nil {
		return nil, err
	}
	ret := make([]float64, len(times))
	for i, t := range times {
		ret[i] = t.User + t.System
	}
	return ret, nil
}
