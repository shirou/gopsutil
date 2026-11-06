// SPDX-License-Identifier: BSD-3-Clause
//go:build dragonfly

package mem

import (
	"context"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// sysctlUint reads a counter that is 4 or 8 bytes wide depending on its kernel type.
func sysctlUint(name string) (uint64, error) {
	buf, err := unix.SysctlRaw(name)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	switch len(buf) {
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf)), nil
	case 8:
		return binary.LittleEndian.Uint64(buf), nil
	}
	return 0, fmt.Errorf("%s: unexpected size %d", name, len(buf))
}

func VirtualMemory() (*VirtualMemoryStat, error) {
	return VirtualMemoryWithContext(context.Background())
}

func VirtualMemoryWithContext(ctx context.Context) (*VirtualMemoryStat, error) {
	pageSize, err := sysctlUint("vm.stats.vm.v_page_size")
	if err != nil {
		return nil, err
	}
	pages := map[string]uint64{}
	for _, name := range []string{"page", "free", "inactive", "active", "wire", "cache"} {
		count, err := sysctlUint("vm.stats.vm.v_" + name + "_count")
		if err != nil {
			return nil, err
		}
		pages[name] = count * pageSize
	}
	buffers, err := sysctlUint("vfs.bufspace")
	if err != nil {
		buffers = 0
	}

	stat := &VirtualMemoryStat{
		Total:    pages["page"],
		Free:     pages["free"],
		Active:   pages["active"],
		Inactive: pages["inactive"],
		Wired:    pages["wire"],
		Cached:   pages["cache"] + pages["inactive"],
		Buffers:  buffers,
	}
	stat.Available = stat.Free + stat.Cached
	reclaimable := stat.Free + stat.Cached + stat.Buffers
	if reclaimable < stat.Total {
		stat.Used = stat.Total - reclaimable
	}
	if stat.Total > 0 {
		stat.UsedPercent = float64(stat.Used) / float64(stat.Total) * 100
	}

	if swap, err := SwapMemoryWithContext(ctx); err == nil {
		stat.SwapTotal = swap.Total
		stat.SwapFree = swap.Free
	}
	return stat, nil
}

func SwapMemory() (*SwapMemoryStat, error) {
	return SwapMemoryWithContext(context.Background())
}

func SwapMemoryWithContext(_ context.Context) (*SwapMemoryStat, error) {
	pageSize, err := sysctlUint("vm.stats.vm.v_page_size")
	if err != nil {
		return nil, err
	}
	total, err := sysctlUint("vm.swap_size")
	if err != nil {
		return nil, err
	}
	free, err := sysctlUint("vm.swap_free")
	if err != nil {
		return nil, err
	}
	stat := &SwapMemoryStat{Total: total * pageSize, Free: free * pageSize}
	stat.Used = stat.Total - stat.Free
	if stat.Total > 0 {
		stat.UsedPercent = float64(stat.Used) / float64(stat.Total) * 100
	}
	return stat, nil
}

func SwapDevices() ([]*SwapDevice, error) {
	return SwapDevicesWithContext(context.Background())
}

func SwapDevicesWithContext(_ context.Context) ([]*SwapDevice, error) {
	return nil, common.ErrNotImplementedError
}
