// SPDX-License-Identifier: BSD-3-Clause
//go:build dragonfly

package net

import (
	"context"
	"encoding/binary"
	"syscall"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// Offsets into struct if_msghdr and struct if_data from net/if.h (RTM_VERSION 7).
// The types in package syscall do not match: they place ifm_addrs before
// ifm_index, which makes syscall.ParseRoutingMessage fail with EINVAL.
const (
	rtmVersion  = 7
	rtmIfInfo   = 0x0e
	ifmDataBase = 16
	ifmSize     = 176 // struct sockaddr_dl with the interface name follows

	ifiIpackets = ifmDataBase + 40
	ifiIerrors  = ifmDataBase + 48
	ifiOpackets = ifmDataBase + 56
	ifiOerrors  = ifmDataBase + 64
	ifiIbytes   = ifmDataBase + 80
	ifiObytes   = ifmDataBase + 88
	ifiIqdrops  = ifmDataBase + 112

	sdlNameLen = 5
	sdlData    = 8
)

// IOCountersWithContext reads the counters from the routing socket sysctl. This
// works for unprivileged users, unlike netstat -i, which reads kernel memory.
func IOCountersWithContext(_ context.Context, pernic bool) ([]IOCountersStat, error) {
	rib, err := syscall.RouteRIB(syscall.NET_RT_IFLIST, 0)
	if err != nil {
		return nil, err
	}

	var counters []IOCountersStat
	for len(rib) >= 4 {
		length := int(binary.LittleEndian.Uint16(rib))
		if length < 4 || length > len(rib) {
			break
		}
		message := rib[:length]
		rib = rib[length:]
		if message[2] != rtmVersion || message[3] != rtmIfInfo || length < ifmSize+sdlData {
			continue
		}

		link := message[ifmSize:]
		nameEnd := sdlData + int(link[sdlNameLen])
		if nameEnd > len(link) {
			continue
		}
		field := func(offset int) uint64 { return binary.LittleEndian.Uint64(message[offset:]) }
		counters = append(counters, IOCountersStat{
			Name:        string(link[sdlData:nameEnd]),
			BytesSent:   field(ifiObytes),
			BytesRecv:   field(ifiIbytes),
			PacketsSent: field(ifiOpackets),
			PacketsRecv: field(ifiIpackets),
			Errin:       field(ifiIerrors),
			Errout:      field(ifiOerrors),
			Dropin:      field(ifiIqdrops),
		})
	}
	if !pernic {
		return getIOCountersAll(counters), nil
	}
	return counters, nil
}

func IOCountersByFileWithContext(ctx context.Context, pernic bool, _ string) ([]IOCountersStat, error) {
	return IOCountersWithContext(ctx, pernic)
}

func FilterCountersWithContext(_ context.Context) ([]FilterStat, error) {
	return nil, common.ErrNotImplementedError
}

func ConntrackStatsWithContext(_ context.Context, _ bool) ([]ConntrackStat, error) {
	return nil, common.ErrNotImplementedError
}

func ProtoCountersWithContext(_ context.Context, _ []string) ([]ProtoCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

// Deprecated: use process.PidsWithContext instead
func PidsWithContext(_ context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
}

func ConnectionsWithContext(_ context.Context, _ string) ([]ConnectionStat, error) {
	return []ConnectionStat{}, common.ErrNotImplementedError
}

func ConnectionsMaxWithContext(ctx context.Context, kind string, maxConn int) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithContext(ctx, kind, 0, maxConn)
}

func ConnectionsWithoutUidsWithContext(ctx context.Context, kind string) ([]ConnectionStat, error) {
	return ConnectionsMaxWithoutUidsWithContext(ctx, kind, 0)
}

func ConnectionsMaxWithoutUidsWithContext(ctx context.Context, kind string, maxConn int) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithoutUidsWithContext(ctx, kind, 0, maxConn)
}

func ConnectionsPidWithoutUidsWithContext(ctx context.Context, kind string, pid int32) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, 0)
}

func ConnectionsPidWithContext(ctx context.Context, kind string, pid int32) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithContext(ctx, kind, pid, 0)
}

func ConnectionsPidMaxWithContext(ctx context.Context, kind string, pid int32, maxConn int) ([]ConnectionStat, error) {
	return connectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, maxConn, false)
}

func ConnectionsPidMaxWithoutUidsWithContext(ctx context.Context, kind string, pid int32, maxConn int) ([]ConnectionStat, error) {
	return connectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, maxConn, true)
}

func connectionsPidMaxWithoutUidsWithContext(_ context.Context, _ string, _ int32, _ int, _ bool) ([]ConnectionStat, error) {
	return []ConnectionStat{}, common.ErrNotImplementedError
}
