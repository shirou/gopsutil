// SPDX-License-Identifier: BSD-3-Clause
//go:build dragonfly

package load

import (
	"context"
	"encoding/binary"
	"errors"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func Avg() (*AvgStat, error) {
	return AvgWithContext(context.Background())
}

// AvgWithContext reads vm.loadavg. Unlike on FreeBSD, struct loadavg holds three
// 64-bit values before the long fscale.
func AvgWithContext(_ context.Context) (*AvgStat, error) {
	buf, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return nil, err
	}
	if len(buf) < 32 {
		return nil, errors.New("vm.loadavg is too short")
	}
	scale := float64(binary.LittleEndian.Uint64(buf[24:]))
	if scale == 0 {
		return nil, errors.New("vm.loadavg has no scale")
	}
	value := func(i int) float64 {
		return float64(binary.LittleEndian.Uint64(buf[i*8:])) / scale
	}
	return &AvgStat{Load1: value(0), Load5: value(1), Load15: value(2)}, nil
}

func Misc() (*MiscStat, error) {
	return MiscWithContext(context.Background())
}

func MiscWithContext(_ context.Context) (*MiscStat, error) {
	return nil, common.ErrNotImplementedError
}
