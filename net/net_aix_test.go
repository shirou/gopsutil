// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package net

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// Abbreviated netstat -s fixture from AIX 7.3 for unit testing.
const testNetstatS = `icmp:
	3 calls to icmp_error
	0 errors not generated because old message was icmp
	Output histogram:
		destination unreachable: 3
	12 messages with bad code fields
	0 messages < minimum length
	0 bad checksums
	0 messages with bad length
	Input histogram:
		destination unreachable: 24
		time exceeded: 34
	0 message responses generated
tcp:
	8864927 packets sent
		4194757 data packets (3255534866 bytes)
		26849 data packets (23271573 bytes) retransmitted
		3174391 ack-only packets (1213424 delayed)
	13907119 packets received
		4921506 acks (for 3255334387 bytes)
		1443623 duplicate acks
	2177 connection requests
	465222 connection accepts
	467064 connections established (including accepts)
	467502 connections closed (including 1913 drops)
udp:
	1886908 datagrams received
	0 incomplete headers
	0 bad checksums
	222742 dropped due to no socket
	1481568 delivered
	2224137 datagrams output
ip:
	16177190 total packets received
	0 bad header checksums
	15571192 packets for this host
	58 packets for unknown/unsupported protocol
	0 packets forwarded
	11912013 packets sent from this host
`

func TestParseNetstatS(t *testing.T) {
	results, err := parseNetstatS(testNetstatS, nil)
	require.NoError(t, err)

	// Should find all 4 protocol sections
	protos := make(map[string]map[string]int64)
	for _, r := range results {
		protos[r.Protocol] = r.Stats
	}
	assert.Contains(t, protos, "tcp")
	assert.Contains(t, protos, "udp")
	assert.Contains(t, protos, "ip")
	assert.Contains(t, protos, "icmp")

	// Verify specific TCP values
	assert.Equal(t, int64(8864927), protos["tcp"]["packetsSent"])
	assert.Equal(t, int64(13907119), protos["tcp"]["packetsReceived"])
	assert.Equal(t, int64(2177), protos["tcp"]["connectionRequests"])
	assert.Equal(t, int64(465222), protos["tcp"]["connectionAccepts"])

	// Verify sub-items (double-tab) are NOT included
	_, hasDataPackets := protos["tcp"]["dataPackets"]
	assert.False(t, hasDataPackets, "sub-items should be skipped")
	_, hasAckOnlyPackets := protos["tcp"]["ackOnlyPackets"]
	assert.False(t, hasAckOnlyPackets, "sub-items should be skipped")

	// Verify UDP values
	assert.Equal(t, int64(1886908), protos["udp"]["datagramsReceived"])
	assert.Equal(t, int64(2224137), protos["udp"]["datagramsOutput"])

	// Verify IP values
	assert.Equal(t, int64(16177190), protos["ip"]["totalPacketsReceived"])
	assert.Equal(t, int64(11912013), protos["ip"]["packetsSentFromThisHost"])

	// Verify ICMP values
	assert.Equal(t, int64(3), protos["icmp"]["callsToIcmpError"])
	assert.Equal(t, int64(12), protos["icmp"]["messagesWithBadCodeFields"])
}

func TestParseNetstatSFiltered(t *testing.T) {
	results, err := parseNetstatS(testNetstatS, []string{"tcp", "udp"})
	require.NoError(t, err)

	protos := make(map[string]bool)
	for _, r := range results {
		protos[r.Protocol] = true
	}
	assert.True(t, protos["tcp"])
	assert.True(t, protos["udp"])
	assert.False(t, protos["ip"])
	assert.False(t, protos["icmp"])
}

func TestParseNetstatSEmpty(t *testing.T) {
	results, err := parseNetstatS("", nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestDescToCamelCase(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"packets sent", "packetsSent"},
		{"total packets received", "totalPacketsReceived"},
		{"data packets (3187866062 bytes)", "dataPackets"},
		{"ack-only packets (1213424 delayed)", "ackOnlyPackets"},
		{"calls to icmp_error", "callsToIcmpError"},
		{"connections established (including accepts)", "connectionsEstablished"},
		{"messages < minimum length", "messagesMinimumLength"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.want, descToCamelCase(tt.desc))
		})
	}
}

func TestProtoCountersWithContext(t *testing.T) {
	ctx := context.Background()
	results, err := ProtoCountersWithContext(ctx, nil)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	for _, r := range results {
		assert.NotEmpty(t, r.Protocol)
		assert.NotEmpty(t, r.Stats)
	}
}
