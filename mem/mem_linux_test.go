// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package mem

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestExVirtualMemory(t *testing.T) {
	ex := NewExLinux()

	v, err := ex.VirtualMemory()
	require.NoError(t, err)

	t.Log(v)
}

var virtualMemoryTests = []struct {
	mockedRootFS string
	stat         *VirtualMemoryStat
	exStat       *ExVirtualMemory
}{
	{
		"intelcorei5", &VirtualMemoryStat{
			Total:          16502300672,
			Available:      11495358464,
			Used:           5006942208,
			UsedPercent:    30.340873721295385,
			Free:           8783491072,
			Active:         4347392000,
			Inactive:       2938834944,
			Wired:          0,
			Laundry:        0,
			Buffers:        212496384,
			Cached:         4069036032,
			WriteBack:      0,
			Dirty:          176128,
			WriteBackTmp:   0,
			Shared:         1222402048,
			Slab:           253771776,
			Sreclaimable:   186470400,
			Sunreclaim:     67301376,
			PageTables:     65241088,
			SwapCached:     0,
			CommitLimit:    16509730816,
			CommittedAS:    12360818688,
			HighTotal:      0,
			HighFree:       0,
			LowTotal:       0,
			LowFree:        0,
			SwapTotal:      8258580480,
			SwapFree:       8258580480,
			Mapped:         1172627456,
			VmallocTotal:   35184372087808,
			VmallocUsed:    0,
			VmallocChunk:   0,
			HugePagesTotal: 0,
			HugePagesFree:  0,
			HugePagesRsvd:  0,
			HugePagesSurp:  0,
			HugePageSize:   2097152,
		},
		&ExVirtualMemory{
			ActiveFile:   1121992 * 1024,
			InactiveFile: 1683344 * 1024,
			ActiveAnon:   3123508 * 1024,
			InactiveAnon: 1186612 * 1024,
			Unevictable:  32 * 1024,
			Percpu:       19136 * 1024,
			KernelStack:  14224 * 1024,
		},
	},
	{
		"issue1002", &VirtualMemoryStat{
			Total:          260579328,
			Available:      215199744,
			Used:           45379584,
			UsedPercent:    17.414882580401773,
			Free:           124506112,
			Active:         108785664,
			Inactive:       8581120,
			Wired:          0,
			Laundry:        0,
			Buffers:        4915200,
			Cached:         96829440,
			WriteBack:      0,
			Dirty:          0,
			WriteBackTmp:   0,
			Shared:         0,
			Slab:           9293824,
			Sreclaimable:   2764800,
			Sunreclaim:     6529024,
			PageTables:     405504,
			SwapCached:     0,
			CommitLimit:    130289664,
			CommittedAS:    25567232,
			HighTotal:      134217728,
			HighFree:       67784704,
			LowTotal:       126361600,
			LowFree:        56721408,
			SwapTotal:      0,
			SwapFree:       0,
			Mapped:         38793216,
			VmallocTotal:   1996488704,
			VmallocUsed:    0,
			VmallocChunk:   0,
			HugePagesTotal: 0,
			HugePagesFree:  0,
			HugePagesRsvd:  0,
			HugePagesSurp:  0,
			HugePageSize:   0,
		},
		&ExVirtualMemory{
			ActiveFile:   88280 * 1024,
			InactiveFile: 8380 * 1024,
			ActiveAnon:   17956 * 1024,
			InactiveAnon: 0,
			Unevictable:  0,
			Percpu:       0,
			KernelStack:  624 * 1024,
		},
	},
	{
		"anonhugepages", &VirtualMemoryStat{
			Total:         260799420 * 1024,
			Available:     127880216 * 1024,
			Free:          119443248 * 1024,
			AnonHugePages: 50409472 * 1024,
			Used:          136109264896,
			UsedPercent:   50.96606579876596,
		},
		&ExVirtualMemory{
			ActiveFile:   0,
			InactiveFile: 0,
			ActiveAnon:   0,
			InactiveAnon: 0,
			Unevictable:  0,
			Percpu:       0,
		},
	},
}

func TestVirtualMemoryLinux(t *testing.T) {
	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			t.Setenv("HOST_PROC", filepath.Join("testdata", "linux", "virtualmemory", tt.mockedRootFS, "proc"))

			stat, err := VirtualMemory()
			if errors.Is(err, common.ErrNotImplementedError) {
				t.Skip("not implemented")
			}
			require.NoError(t, err)
			assert.Truef(t, reflect.DeepEqual(stat, tt.stat), "got: %+v\nwant: %+v", stat, tt.stat)
		})
	}
}

func TestExVirtualMemoryLinux(t *testing.T) {
	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			t.Setenv("HOST_PROC", filepath.Join("testdata", "linux", "virtualmemory", tt.mockedRootFS, "proc"))

			ex := NewExLinux()
			exStat, err := ex.VirtualMemory()
			if errors.Is(err, common.ErrNotImplementedError) {
				t.Skip("not implemented")
			}
			require.NoError(t, err)
			assert.Truef(t, reflect.DeepEqual(exStat, tt.exStat), "got: %+v\nwant: %+v", exStat, tt.exStat)
		})
	}
}

const validFile = `Filename				Type		Size		Used		Priority
/dev/dm-2                               partition	67022844	490788		-2
/swapfile                               file		2		1		-3
`

const invalidFile = `INVALID				Type		Size		Used		Priority
/dev/dm-2                               partition	67022844	490788		-2
/swapfile                               file		1048572		0		-3
`

func TestParseSwapsFile_ValidFile(t *testing.T) {
	stats, err := parseSwapsFile(context.Background(), strings.NewReader(validFile))
	require.NoError(t, err)

	assert.Equal(t, SwapDevice{
		Name:      "/dev/dm-2",
		UsedBytes: 502566912,
		FreeBytes: 68128825344,
	}, *stats[0])

	assert.Equal(t, SwapDevice{
		Name:      "/swapfile",
		UsedBytes: 1024,
		FreeBytes: 1024,
	}, *stats[1])
}

func TestParseSwapsFile_InvalidFile(t *testing.T) {
	_, err := parseSwapsFile(context.Background(), strings.NewReader(invalidFile))
	assert.Error(t, err)
}

func TestParseSwapsFile_EmptyFile(t *testing.T) {
	_, err := parseSwapsFile(context.Background(), strings.NewReader(""))
	assert.Error(t, err)
}
