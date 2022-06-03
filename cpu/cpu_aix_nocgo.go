//go:build aix && !cgo
// +build aix,!cgo

package cpu

import (
	"context"
	"strings"
	"strconv"
)

func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	return []TimesStat{}, common.ErrNotImplementedError
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	return []InfoStat{}, common.ErrNotImplementedError
}

func CountsWithContext(ctx context.Context, logical bool) (int, error) {
	prtConfOut, err := invoke.CommandWithContext(ctx, "prtconf")
	if err != nil {
		return 0, fmt.Errorf("cannot execute prtconf: %s", err)
	}
	for _, line := range strings.Split(string(prtConfOut), "\n") {
		if parts := strings.Split(line, ": "); len(parts) < 2 {
			continue
		}			
		if parts[0] == "Number Of Processors" {
			if ncpu, err := strconv.Atoi(parts[1]); err == nil {
				return ncpu, nil
			}
		}
	}
	return 0, fmt.Errorf("number of processors not found")
}
