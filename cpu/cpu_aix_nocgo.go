//go:build aix && !cgo
// +build aix,!cgo

package cpu

import (
	"context"
	"regexp"
	"strings"
	"strconv"

	"github.com/shirou/gopsutil/v3/internal/common"
)

var whiteSpaces = regexp.MustCompile(`\s+`)

func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	if percpu {
		return []TimesStat{}, common.ErrNotImplementedError
	} else {
		out, err := invoke.CommandWithContext(ctx, "sar", "-u", "10", "1")
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) < 5 {
			return []TimesStat{}, common.ErrNotImplementedError
		}

		ret := TimesStat{CPU: "cpu-total"}
		h := whiteSpaces.Split(lines[len(lines)-3], -1)	// headers
		v := whiteSpaces.Split(lines[len(lines)-2], -1)	// values
		for i, header := range h {
			if t, err := strconv.ParseUint(v[i], 10, 64); err == nil {
				switch header {
					case `%usr`:
						ret.User = float64(t)
					case `%sys`:
						ret.System = float64(t)
					case `%wio`:
						ret.Iowait = float64(t)
					case `%idle`:
						ret.Idle = float64(t)
				}
			}
		}

		return []TimesStat{ret}, nil
	}
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	out, err := invoke.CommandWithContext(ctx, "prtconf")
	if err != nil {
		return nil, err
	}

	ret := InfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Number Of Processors:") {
			p := whiteSpaces.Split(line, 4)
			if len(p) > 3 {
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					ret.Cores = int32(t)
				}
			}
		} else if strings.HasPrefix(line, "Processor Clock Speed:") {
			p := whiteSpaces.Split(line, 5)
			if len(p) > 4 {
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					switch strings.ToUpper(p[4]) {
					case "MHZ":
						ret.Mhz = float64(t)
					case "GHZ":
						ret.Mhz = float64(t) * 1000.0
					case "KHZ":
						ret.Mhz = float64(t) / 1000.0
					default:
						ret.Mhz = float64(t)
					}
				}
			}
			break
		}
	}
	return []InfoStat{ret}, nil
}

func CountsWithContext(ctx context.Context, logical bool) (int, error) {
	info, err := InfoWithContext(ctx)
	if err == nil {
		return int(info[0].Cores), nil
	}
	return 0, err
}
