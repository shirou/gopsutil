// SPDX-License-Identifier: BSD-3-Clause
// Tracing network packages with "gopacket"  package
package net

import (
	"context"
	"sync"
	"time"
)

// Network IO counters for a process
type ProcNetStat struct {
	Pid         int32          `json:"pid"`       // process PID
	NetCounters IOCountersStat `json:"netCounts"` // process network counters
}

// For unit test mocking
type connProviderF func(_ context.Context, _ string) ([]ConnectionStat, error)

var (
	ProcConnMap      map[Addr]ProcNetStat
	void             = struct{}{}
	inactiveStatuses = map[string]struct{}{"CLOSED": void, "CLOSE": void, "TIME_WAIT": void, "DELETE": void}
	watchLock        sync.RWMutex
	errChannel       chan error
)

// Start collecting information on open ports and network traffic
func StartTracing(ctx context.Context, kind string, intvl time.Duration) chan error {
	if ProcConnMap != nil {
		panic("Repeated capturing of process NET I/O is not supported")
	}

	ProcConnMap = make(map[Addr]ProcNetStat)
	errChannel = make(chan error)

	go pollNetStat(ctx, kind, intvl)
	tracePackets(ctx, kind)

	return errChannel
}

func GetProcConnStat() map[Addr]ProcNetStat {
	watchLock.RLock()
	defer watchLock.RUnlock()

	return ProcConnMap
}

func pollNetStat(ctx context.Context, kind string, intvl time.Duration) {
	watchTicker := time.NewTicker(intvl)

	for range watchTicker.C {
		select {
		case <-ctx.Done():
			return
		default:
			updateTable(ctx, kind, ConnectionsWithContext)
		}
	}
}

// testable
func updateTable(ctx context.Context, kind string, connProvider connProviderF) {
	watchLock.Lock()
	defer watchLock.Unlock()

	conns, err := connProvider(ctx, kind)
	if err != nil {
		errChannel <- err
		return
	}

	portPidMap := make(map[Addr]int32)
	for _, conn := range conns {
		if _, ok := inactiveStatuses[conn.Status]; !ok {
			portPidMap[conn.Laddr] = conn.Pid
		}
	}

	// remove outdated entries
	for a := range ProcConnMap {
		if _, ok := portPidMap[a]; !ok {
			delete(ProcConnMap, a)
		}
	}
	// add new entries
	for a, p := range portPidMap {
		if _, ok := ProcConnMap[a]; !ok {
			ProcConnMap[a] = ProcNetStat{Pid: p, NetCounters: IOCountersStat{}}
		}
	}
}
