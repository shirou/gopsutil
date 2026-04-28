// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package net

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNetstatS(t *testing.T) {
	data, err := os.ReadFile("testdata/aix/netstat_s.txt")
	require.NoError(t, err)

	stats, err := parseNetstatS(string(data), nil)
	require.NoError(t, err)

	protos := map[string]ProtoCountersStat{}
	for _, s := range stats {
		protos[s.Protocol] = s
	}

	var gotProtocols []string
	for p := range protos {
		gotProtocols = append(gotProtocols, p)
	}
	assert.ElementsMatch(t, []string{"tcp", "udp", "ip", "ipv6"}, gotProtocols)

	t.Run("tcp", func(t *testing.T) {
		tcp, ok := protos["tcp"]
		require.True(t, ok, "tcp section not found")
		assert.Equal(t, map[string]int64{
			"OutSegs":      1001,
			"InSegs":       2002,
			"RetransSegs":  303,
			"ActiveOpens":  404,
			"PassiveOpens": 505,
			"AttemptFails": 606,
			"InCsumErrors": 707,
			"InErrs":       1603, // 801 (bad header offset) + 802 (packet too short)
		}, tcp.Stats)
	})

	t.Run("udp", func(t *testing.T) {
		udp, ok := protos["udp"]
		require.True(t, ok, "udp section not found")
		assert.Equal(t, map[string]int64{
			"InDatagrams":  1100,
			"OutDatagrams": 2200,
			"NoPorts":      7700, // 3300 (unicast) + 4400 (broadcast)
			"InCsumErrors": 5500,
			"RcvbufErrors": 6600,
			"InErrors":     330, // 110 (incomplete headers) + 220 (bad data length)
		}, udp.Stats)
	})

	t.Run("ip", func(t *testing.T) {
		ip, ok := protos["ip"]
		require.True(t, ok, "ip section not found")
		assert.Equal(t, map[string]int64{
			"InReceives":      50000,
			"InDelivers":      40000,
			"OutRequests":     30000,
			"InHdrErrors":     308, // 11+22+33+44+55+66+77 (seven header-error lines)
			"InUnknownProtos": 9000,
			"ForwDatagrams":   100,
			"ReasmOKs":        200,
			"ReasmReqds":      900,
			"ReasmFails":      300,
			"InDiscards":      500,
			"FragCreates":     600,
			"OutNoRoutes":     700,
			"OutDiscards":     800,
		}, ip.Stats)
	})

	t.Run("ipv6", func(t *testing.T) {
		ipv6, ok := protos["ipv6"]
		require.True(t, ok, "ipv6 section not found")
		assert.Equal(t, map[string]int64{
			"InReceives":      60000,
			"InDelivers":      55000,
			"OutRequests":     45000,
			"ForwDatagrams":   110,
			"ReasmOKs":        220,
			"ReasmReqds":      990,
			"ReasmFails":      330,
			"InDiscards":      651, // 550 (fragments) + 101 (input no memory)
			"FragCreates":     660,
			"OutNoRoutes":     770,
			"InHdrErrors":     110, // 11+22+33+44 (four header-error lines)
			"InUnknownProtos": 880,
			"OutDiscards":     777, // 333 (no bufs) + 444 (no memory)
		}, ipv6.Stats)
	})
}

func TestParseNetstatSFiltered(t *testing.T) {
	data, err := os.ReadFile("testdata/aix/netstat_s.txt")
	require.NoError(t, err)

	stats, err := parseNetstatS(string(data), []string{"tcp", "udp"})
	require.NoError(t, err)

	protos := map[string]ProtoCountersStat{}
	for _, s := range stats {
		protos[s.Protocol] = s
	}
	assert.Contains(t, protos, "tcp")
	assert.Contains(t, protos, "udp")
	assert.NotContains(t, protos, "ip")
	assert.NotContains(t, protos, "ipv6")
}

func TestNormaliseNetstatDesc(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"packets sent", "packets sent"},
		{"data packets (6302116893 bytes) retransmitted", "data packets retransmitted"},
		{"823508 segments updated rtt (of 120622 attempts)", "823508 segments updated rtt"},
		{"connections closed (including 74 drops)", "connections closed"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, normaliseNetstatDesc(tc.input), "input: %q", tc.input)
	}
}
