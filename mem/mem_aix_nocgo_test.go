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
