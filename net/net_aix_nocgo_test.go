// SPDX-License-Identifier: BSD-3-Clause
//go:build aix && !cgo

package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Captured from a real AIX 7.3 system: entstat en0
const testEntstatOutput = `-------------------------------------------------------------
ETHERNET STATISTICS (en0) :
Device Type: Virtual I/O Ethernet Adapter (l-lan)
Hardware Address: 4e:12:19:56:b5:14
Elapsed Time: 22 days 6 hours 6 minutes 25 seconds

Transmit Statistics:                          Receive Statistics:
--------------------                          -------------------
Packets: 2901731                              Packets: 7163285
Bytes: 699977762                              Bytes: 1084168420
Interrupts: 0                                 Interrupts: 6977209
Transmit Errors: 0                            Receive Errors: 0
Packets Dropped: 0                            Packets Dropped: 0
                                              Bad Packets: 0
Max Packets on S/W Transmit Queue: 0
S/W Transmit Queue Overflow: 0
Current S/W+H/W Transmit Queue Length: 0

Broadcast Packets: 349                        Broadcast Packets: 4144143
Multicast Packets: 2                          Multicast Packets: 0
No Carrier Sense: 0                           CRC Errors: 0
DMA Underrun: 0                               DMA Overrun: 0
Lost CTS Errors: 0                            Alignment Errors: 0
Max Collision Errors: 0                       No Resource Errors: 0
Late Collision Errors: 0                      Receive Collision Errors: 0
Deferred: 0                                   Packet Too Short Errors: 0
SQE Test: 0                                   Packet Too Long Errors: 0
Timeout Errors: 0                             Packets Discarded by Adapter: 0
Single Collision Count: 0                     Receiver Start Count: 0
Multiple Collision Count: 0
Current HW Transmit Queue Length: 0

General Statistics:
-------------------
No mbuf Errors: 0
Adapter Reset Count: 0
Adapter Data Rate: 20000
Driver Flags: Up Broadcast Running
	Simplex 64BitSupport ChecksumOffload
	LargeSend DataRateSet PlatformTSO
	VIOENT IPV6_LSO IPV6_CSO
`

func TestParseEntstat(t *testing.T) {
	bytesSent, bytesRecv := parseEntstat(testEntstatOutput)
	assert.Equal(t, uint64(699977762), bytesSent)
	assert.Equal(t, uint64(1084168420), bytesRecv)
}

func TestParseEntstatEmpty(t *testing.T) {
	bytesSent, bytesRecv := parseEntstat("")
	assert.Equal(t, uint64(0), bytesSent)
	assert.Equal(t, uint64(0), bytesRecv)
}

func TestParseEntstatNoBytesLine(t *testing.T) {
	// Output with no "Bytes:" line at all
	bytesSent, bytesRecv := parseEntstat("some random output\nno bytes here\n")
	assert.Equal(t, uint64(0), bytesSent)
	assert.Equal(t, uint64(0), bytesRecv)
}

func TestParseEntstatSingleColumn(t *testing.T) {
	// Only transmit bytes, no receive column
	bytesSent, bytesRecv := parseEntstat("Bytes: 12345\n")
	assert.Equal(t, uint64(12345), bytesSent)
	assert.Equal(t, uint64(0), bytesRecv)
}

// Captured from a real AIX 7.3 system: netstat -idn
// The parseNetstatI output format has been stable across AIX versions.
const testNetstatIdnOutput = `Name   Mtu   Network     Address                 Ipkts     Ierrs        Opkts     Oerrs  Coll  Drop
en0    1500  link#2      4e.12.19.56.b5.14          7167418     0          2902787     0     0     0
en0    1500  192.168.240 192.168.242.122           7167418     0          2902787     0     0     0
lo0    16896 link#1                                1986929     0          1986929     0     0     0
lo0    16896 127         127.0.0.1                 1986929     0          1986929     0     0     0
lo0    16896 ::1%1                                 1986929     0          1986929     0     0     0
`

func TestParseNetstatI(t *testing.T) {
	result, err := parseNetstatI(testNetstatIdnOutput)
	require.NoError(t, err)

	// Should deduplicate — en0 appears twice, lo0 three times
	require.Len(t, result, 2)

	// en0 (first occurrence: link#2 line with Address field)
	assert.Equal(t, "en0", result[0].Name)
	assert.Equal(t, uint64(7167418), result[0].PacketsRecv)
	assert.Equal(t, uint64(0), result[0].Errin)
	assert.Equal(t, uint64(2902787), result[0].PacketsSent)
	assert.Equal(t, uint64(0), result[0].Errout)
	assert.Equal(t, uint64(0), result[0].Dropout)

	// lo0 (first occurrence: link#1 line without Address field)
	assert.Equal(t, "lo0", result[1].Name)
	assert.Equal(t, uint64(1986929), result[1].PacketsRecv)
	assert.Equal(t, uint64(0), result[1].Errin)
	assert.Equal(t, uint64(1986929), result[1].PacketsSent)
	assert.Equal(t, uint64(0), result[1].Errout)
	assert.Equal(t, uint64(0), result[1].Dropout)
}

func TestParseNetstatIInvalidHeader(t *testing.T) {
	_, err := parseNetstatI("not a valid header\n")
	assert.Error(t, err)
}

func TestParseNetstatIEmpty(t *testing.T) {
	_, err := parseNetstatI("Name   Mtu   Network\n")
	require.NoError(t, err)
}
