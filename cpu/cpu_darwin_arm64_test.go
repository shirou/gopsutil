// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && arm64

package cpu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMhzFromRawClock(t *testing.T) {
	tests := []struct {
		name  string
		model string
		raw   uint32
		want  float64
	}{
		{name: "M1 Hz", model: "Apple M1", raw: 3_204_000_000, want: 3204},
		{name: "M3 Hz", model: "Apple M3 Max", raw: 2_748_000_000, want: 2748},
		{name: "unknown model stays Hz", model: "", raw: 3_204_000_000, want: 3204},
		{name: "M4 kHz", model: "Apple M4 Pro", raw: 4_512_000, want: 4512},
		{name: "M5 kHz", model: "Apple M5 Pro", raw: 4_608_000, want: 4608},
		{name: "M5 Max efficiency kHz", model: "Apple M5 Max", raw: 4_380_000, want: 4380},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, mhzFromRawClock(tt.raw, tt.model), 0.001)
		})
	}
}

func TestParsePCoreHz(t *testing.T) {
	tests := []struct {
		name    string
		buf     []byte
		want    uint32
		wantErr bool
	}{
		{name: "nil", buf: nil, wantErr: true},
		{name: "empty", buf: []byte{}, wantErr: true},
		{name: "short", buf: []byte{1, 2, 3, 4, 5, 6, 7}, wantErr: true},
		{
			name: "exactly two words",
			// 3504000000 Hz = 0xD0DACC00, then a trailing voltage word
			buf:  []byte{0x00, 0xCC, 0xDA, 0xD0, 0x10, 0x03, 0x00, 0x00},
			want: 3_504_000_000,
		},
		{
			name: "takes the second-to-last word",
			buf: []byte{
				0x00, 0x2d, 0x31, 0x01, 0x00, 0x00, 0x00, 0x00, // 20000000 Hz, voltage
				0x00, 0xCC, 0xDA, 0xD0, 0x10, 0x03, 0x00, 0x00, // 3504000000 Hz, voltage
			},
			want: 3_504_000_000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePCoreHz(tt.buf)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
