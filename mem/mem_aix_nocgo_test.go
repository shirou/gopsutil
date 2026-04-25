// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLspsOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*SwapDevice
	}{
		{
			name: "MB size",
			input: `Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
hd6             hdisk6            rootvg         512MB     3     yes     yes      lv       0
`,
			expected: []*SwapDevice{
				{
					Name:      "hd6",
					UsedBytes: 512 * 1024 * 1024 * 3 / 100,
					FreeBytes: 512*1024*1024 - 512*1024*1024*3/100,
				},
			},
		},
		{
			name: "GB size",
			input: `Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
paging00        hdisk0            rootvg           4GB    10     yes     yes      lv       0
`,
			expected: []*SwapDevice{
				{
					Name:      "paging00",
					UsedBytes: 4 * 1024 * 1024 * 1024 * 10 / 100,
					FreeBytes: 4*1024*1024*1024 - 4*1024*1024*1024*10/100,
				},
			},
		},
		{
			name: "NaNQ percent used (NFS paging space)",
			input: `Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
nfspg           -                 -              256MB  NaNQ     yes     yes     nfs       0
`,
			expected: []*SwapDevice{
				{
					Name:      "nfspg",
					UsedBytes: 0,
					FreeBytes: 256 * 1024 * 1024,
				},
			},
		},
		{
			name: "multiple devices mixed units",
			input: `Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
hd6             hdisk0            rootvg         512MB     5     yes     yes      lv       0
paging01        hdisk1            datavg           2GB    20     yes     yes      lv       0
`,
			expected: []*SwapDevice{
				{
					Name:      "hd6",
					UsedBytes: 512 * 1024 * 1024 * 5 / 100,
					FreeBytes: 512*1024*1024 - 512*1024*1024*5/100,
				},
				{
					Name:      "paging01",
					UsedBytes: 2 * 1024 * 1024 * 1024 * 20 / 100,
					FreeBytes: 2*1024*1024*1024 - 2*1024*1024*1024*20/100,
				},
			},
		},
		{
			name:     "empty output",
			input:    "",
			expected: nil,
		},
		{
			name: "header only",
			input: `Page Space      Physical Volume   Volume Group    Size %Used   Active    Auto    Type   Chksum
`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := parseLspsOutput(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, devices)
		})
	}
}

func TestParseLspsSize(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		wantErr  bool
	}{
		{"512MB", 512 * 1024 * 1024, false},
		{"4GB", 4 * 1024 * 1024 * 1024, false},
		{"1TB", 1024 * 1024 * 1024 * 1024, false},
		{"0MB", 0, false},
		{"notasize", 0, true},
		{"MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseLspsSize(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Sample svmon -G output (raw pages, 4KB page size)
const testSVMonPages = `               size       inuse        free         pin     virtual   mmode
memory      1048576      544457      504119      383555      425029     Ded
pg space     131072        3644

               work        pers        clnt       other
pin          275224           0        4203      104128
in use       425029           0      119428
`

// Sample svmon -G -O unit=KB output
const testSVMonKB = `Unit: KB
--------------------------------------------------------------------------------------
               size       inuse        free         pin     virtual  available   mmode
memory      4194304     2177848     2016456     1534220     1700132    2126476     Ded
pg space     524288       14576
`

func TestParseSVMonPages(t *testing.T) {
	pagesize := uint64(4096)
	vmem, swap := parseSVMonPages(testSVMonPages, pagesize, true)

	// memory line: 1048576 pages * 4096 = 4294967296 bytes (4 GB)
	assert.Equal(t, uint64(1048576*4096), vmem.Total)
	assert.Equal(t, uint64(544457*4096), vmem.Used)
	assert.Equal(t, uint64(504119*4096), vmem.Free)
	assert.InDelta(t, 51.92, vmem.UsedPercent, 0.1)
	// Available is not in raw page output — stays zero
	assert.Equal(t, uint64(0), vmem.Available)

	// pg space: 131072 pages * 4096 = 536870912 bytes (512 MB)
	assert.Equal(t, uint64(131072*4096), swap.Total)
	assert.Equal(t, uint64(3644*4096), swap.Used)
	assert.Equal(t, swap.Total-swap.Used, swap.Free)
	assert.InDelta(t, 2.78, swap.UsedPercent, 0.1)
}

func TestParseSVMonPagesSwapOnly(t *testing.T) {
	pagesize := uint64(4096)
	vmem, swap := parseSVMonPages(testSVMonPages, pagesize, false)

	// Memory fields should be zero when parseMemory=false
	assert.Equal(t, uint64(0), vmem.Total)

	// Swap should still be parsed
	assert.Equal(t, uint64(131072*4096), swap.Total)
	assert.Equal(t, uint64(3644*4096), swap.Used)
}

func TestParseSVMonAvailable(t *testing.T) {
	available := parseSVMonAvailable(testSVMonKB)
	// 2126476 KB * 1024 = 2177511424 bytes
	assert.Equal(t, uint64(2126476*1024), available)
}

func TestParseSVMonAvailableEmpty(t *testing.T) {
	available := parseSVMonAvailable("")
	assert.Equal(t, uint64(0), available)
}
