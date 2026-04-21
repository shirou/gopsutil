// SPDX-License-Identifier: BSD-3-Clause
//go:build netbsd

package net

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNetstat(t *testing.T) {
	tests := []struct {
		file string
		mode string
		// expected values per interface name
		want map[string]IOCountersStat
	}{
		{
			file: "netstat_inb.txt",
			mode: "inb",
			want: map[string]IOCountersStat{
				"vioif": {Name: "vioif", BytesRecv: 3716508243, BytesSent: 2240547599},
				"lo0":   {Name: "lo0", BytesRecv: 41764800, BytesSent: 41764800},
			},
		},
		{
			file: "netstat_ind.txt",
			mode: "ind",
			want: map[string]IOCountersStat{
				"vioif": {Name: "vioif", PacketsRecv: 11524361, PacketsSent: 18735351},
				"lo0":   {Name: "lo0", PacketsRecv: 835296, PacketsSent: 835296},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "netbsd", tt.file))
			require.NoErrorf(t, err, "reading %s", tt.file)

			iocs := make(map[string]IOCountersStat)
			err = parseNetstat(string(data), tt.mode, iocs)
			require.NoErrorf(t, err, "parseNetstat(%s)", tt.file)

			require.Len(t, iocs, len(tt.want), "unexpected number of interfaces")

			for name, want := range tt.want {
				got, ok := iocs[name]
				require.Truef(t, ok, "interface %q not found in parsed output", name)
				assert.Equal(t, want.Name, got.Name)
				if tt.mode == "inb" {
					assert.Equal(t, want.BytesRecv, got.BytesRecv, "%s BytesRecv", name)
					assert.Equal(t, want.BytesSent, got.BytesSent, "%s BytesSent", name)
				} else {
					assert.Equal(t, want.PacketsRecv, got.PacketsRecv, "%s PacketsRecv", name)
					assert.Equal(t, want.PacketsSent, got.PacketsSent, "%s PacketsSent", name)
				}
			}
		})
	}
}
