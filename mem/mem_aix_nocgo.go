// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package mem

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func VirtualMemoryWithContext(ctx context.Context) (*VirtualMemoryStat, error) {
	vmem, swap, err := callSVMon(ctx, true)
	if err != nil {
		return nil, err
	}
	if vmem.Total == 0 {
		return nil, common.ErrNotImplementedError
	}
	vmem.SwapTotal = swap.Total
	vmem.SwapFree = swap.Free
	return vmem, nil
}

func SwapMemoryWithContext(ctx context.Context) (*SwapMemoryStat, error) {
	_, swap, err := callSVMon(ctx, false)
	if err != nil {
		return nil, err
	}
	if swap.Total == 0 {
		return nil, common.ErrNotImplementedError
	}
	return swap, nil
}

func SwapDevicesWithContext(ctx context.Context) ([]*SwapDevice, error) {
	out, err := invoke.CommandWithContext(ctx, "lsps", "-a")
	if err != nil {
		return nil, err
	}
	return parseLspsOutput(string(out))
}

// parseLspsOutput parses the output of "lsps -a" into SwapDevice entries.
//
// lsps -a output format:
//
//	Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
//	hd6             hdisk6            rootvg         512MB     3     yes     yes      lv       0
func parseLspsOutput(output string) ([]*SwapDevice, error) {
	var ret []*SwapDevice
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		// Skip header line
		if fields[0] == "Page" {
			continue
		}

		totalBytes, err := parseLspsSize(fields[3])
		if err != nil {
			continue
		}

		// %Used may be "NaNQ" for NFS paging spaces — treat as 0
		pctUsed, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			pctUsed = 0
		}

		usedBytes := totalBytes * pctUsed / 100
		ret = append(ret, &SwapDevice{
			Name:      fields[0],
			UsedBytes: usedBytes,
			FreeBytes: totalBytes - usedBytes,
		})
	}
	return ret, nil
}

// parseLspsSize parses a size string from lsps output (e.g., "512MB", "4GB", "1TB").
func parseLspsSize(s string) (uint64, error) {
	units := []struct {
		suffix     string
		multiplier uint64
	}{
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			val, err := strconv.ParseUint(strings.TrimSuffix(s, u.suffix), 10, 64)
			if err != nil {
				return 0, err
			}
			return val * u.multiplier, nil
		}
	}
	return 0, fmt.Errorf("unsupported size unit in %q", s)
}

func callSVMon(ctx context.Context, virt bool) (*VirtualMemoryStat, *SwapMemoryStat, error) {
	out, err := invoke.CommandWithContext(ctx, "svmon", "-G", "-O", "unit=KB")
	if err != nil {
		return nil, nil, err
	}

	vmem := &VirtualMemoryStat{}
	swap := &SwapMemoryStat{}
	for _, line := range strings.Split(string(out), "\n") {
		if virt && strings.HasPrefix(line, "memory") {
			// memory  size  inuse  free  pin  virtual  available  mmode
			// Fields: 0     1      2     3    4        5          6        7
			p := strings.Fields(line)
			if len(p) > 6 {
				if t, err := strconv.ParseUint(p[1], 10, 64); err == nil {
					vmem.Total = t * 1024
				}
				if t, err := strconv.ParseUint(p[2], 10, 64); err == nil {
					vmem.Used = t * 1024
					if vmem.Total > 0 {
						vmem.UsedPercent = 100 * float64(vmem.Used) / float64(vmem.Total)
					}
				}
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					vmem.Free = t * 1024
				}
				if t, err := strconv.ParseUint(p[6], 10, 64); err == nil {
					vmem.Available = t * 1024
				}
			}
		} else if strings.HasPrefix(line, "pg space") {
			// "pg space" splits as: pg space total used
			// Fields:               0  1     2     3
			p := strings.Fields(line)
			if len(p) > 3 {
				if t, err := strconv.ParseUint(p[2], 10, 64); err == nil {
					swap.Total = t * 1024
				}
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					swapUsed := t * 1024
					swap.Used = swapUsed
					swap.Free = swap.Total - swapUsed
					if swap.Total > 0 {
						swap.UsedPercent = 100 * float64(swap.Used) / float64(swap.Total)
					}
				}
			}
			break
		}
	}
	return vmem, swap, nil
}
