// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package cpu

import (
	"context"
)

func Times(percpu bool) ([]TimesStat, error) {
	return TimesWithContext(context.Background(), percpu)
}

func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

// aixPercent returns CPU busy percentages directly from TimesWithContext,
// which on AIX already contains instantaneous percentage values.
func aixPercent(ctx context.Context, percpu bool) ([]float64, error) {
	times, err := TimesWithContext(ctx, percpu)
	if err != nil {
		return nil, err
	}
	ret := make([]float64, len(times))
	for i, t := range times {
		ret[i] = t.User + t.System + t.Iowait
	}
	return ret, nil
}
