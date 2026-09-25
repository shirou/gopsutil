// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin && arm64

package cpu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
