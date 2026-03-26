// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package mem

import (
	"context"
	"fmt"
	"os"
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

// parseSVMonPages parses raw "svmon -G" output (page counts) into memory
// and swap statistics. Using raw page counts preserves full precision;
// the page size is provided by the caller (typically os.Getpagesize()).
//
// svmon -G output format:
//
//	               size       inuse        free         pin     virtual   mmode
//	memory      1048576      544457      504119      383555      425029     Ded
//	pg space     131072        3644
func parseSVMonPages(output string, pagesize uint64, parseMemory bool) (*VirtualMemoryStat, *SwapMemoryStat) {
	vmem := &VirtualMemoryStat{}
	swap := &SwapMemoryStat{}
	for _, line := range strings.Split(output, "\n") {
		if parseMemory && strings.HasPrefix(line, "memory") {
			p := strings.Fields(line)
			if len(p) > 5 {
				if t, err := strconv.ParseUint(p[1], 10, 64); err == nil {
					vmem.Total = t * pagesize
				}
				if t, err := strconv.ParseUint(p[2], 10, 64); err == nil {
					vmem.Used = t * pagesize
					if vmem.Total > 0 {
						vmem.UsedPercent = 100 * float64(vmem.Used) / float64(vmem.Total)
					}
				}
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					vmem.Free = t * pagesize
				}
			}
		} else if strings.HasPrefix(line, "pg space") {
			// "pg space" is two words, so fields split as: pg space total used
			p := strings.Fields(line)
			if len(p) > 3 {
				if t, err := strconv.ParseUint(p[2], 10, 64); err == nil {
					swap.Total = t * pagesize
				}
				if t, err := strconv.ParseUint(p[3], 10, 64); err == nil {
					swapUsed := t * pagesize
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
	return vmem, swap
}

// parseSVMonAvailable extracts the Available memory value from
// "svmon -G -O unit=KB" output. The available column only exists in
// this format — it is a kernel-computed value that includes free pages
// plus reclaimable cache and cannot be derived from raw page counts.
// Returns the value in bytes.
//
// svmon -G -O unit=KB output format:
//
//	               size       inuse        free         pin     virtual  available   mmode
//	memory      4194304     2177848     2016456     1534220     1700132    2126476     Ded
func parseSVMonAvailable(output string) uint64 {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "memory") {
			p := strings.Fields(line)
			if len(p) > 6 {
				if t, err := strconv.ParseUint(p[6], 10, 64); err == nil {
					return t * 1024
				}
			}
			break
		}
	}
	return 0
}

// callSVMon collects memory and swap statistics from svmon.
//
// Two separate svmon calls are used:
//  1. "svmon -G" (raw 4KB pages) — provides size, inuse, free, pin, virtual
//     for both memory and paging space. Using raw page counts preserves full
//     precision; the page size comes from os.Getpagesize().
//  2. "svmon -G -O unit=KB" — only used when virt=true, to read the
//     "available" column which does not exist in the raw page output.
//     Available memory includes free pages plus reclaimable cache and is
//     computed internally by the AIX kernel.
//
// This two-call approach avoids the precision loss from -O unit=KB (which
// truncates sub-KB fractions) while still providing the real available value.
func callSVMon(ctx context.Context, virt bool) (*VirtualMemoryStat, *SwapMemoryStat, error) {
	pagesize := uint64(os.Getpagesize())

	out, err := invoke.CommandWithContext(ctx, "svmon", "-G")
	if err != nil {
		return nil, nil, err
	}

	vmem, swap := parseSVMonPages(string(out), pagesize, virt)

	// Separate call for the available column, which only exists in unit=KB output.
	if virt {
		kbOut, err := invoke.CommandWithContext(ctx, "svmon", "-G", "-O", "unit=KB")
		if err == nil {
			vmem.Available = parseSVMonAvailable(string(kbOut))
		}
	}

	return vmem, swap, nil
}
