// SPDX-License-Identifier: BSD-3-Clause
package net

import (
	"context"
	"sync"
	"time"
)

// Network IO counters for a process
type ProcNetStat struct {
	Pid         int32          `json:"pid"`        // process PID
	NetCounters IOCountersStat `json:"netCounts"`  // process network counters
	RemoteAddr  Addr           `json:"remoteAddr"` // remote address
	LastUpdate  time.Time      `json:"lastUpd"`    // last updated
}

// For unit test mocking
type connProviderF func(_ context.Context, _ string) ([]ConnectionStat, error)

var (
	ProcConnMap      map[Addr]*ProcNetStat
	watchLock        sync.RWMutex
	errChan          chan error
	void             = struct{}{}
	inactiveStatuses = map[string]struct{}{"CLOSED": void, "CLOSE": void, "TIME_WAIT": void, "DELETE": void}
)

// Start collecting information on open ports and network traffic with Go-Pcap
func StartTracing(ctx context.Context, kind string, intvl time.Duration) chan error {
	if ProcConnMap != nil {
		panic("Repeated capturing of process NET I/O is not supported")
	}

	ProcConnMap = make(map[Addr]*ProcNetStat)
	errChan = make(chan error)

	go pollNetStat(ctx, kind, intvl)
	tracePackets(ctx, kind)

	return errChan
}

func GetProcConnStat() map[Addr]ProcNetStat {
	watchLock.RLock()
	defer watchLock.RUnlock()

	return copyMap(ProcConnMap)
}

// Deep copy!
func copyMap[K comparable, V any](orgMap map[K]*V) map[K]V {
	newMap := make(map[K]V)
	for key, value := range orgMap {
		newMap[key] = *value
	}
	return newMap
}

func pollNetStat(ctx context.Context, kind string, intvl time.Duration) {
	watchTicker := time.NewTicker(intvl)

	for range watchTicker.C {
		select {
		case <-ctx.Done():
			return
		default:
			updateTable(ctx, kind, ConnectionsWithContext, 2*intvl)
		}
	}
}

// testable
func updateTable(ctx context.Context, kind string, connProvider connProviderF, expiry time.Duration) {
	watchLock.Lock()
	defer watchLock.Unlock()

	conns, err := connProvider(ctx, kind)
	if err != nil {
		errChan <- err
		return
	}

	portPidMap := make(map[Addr]ConnectionStat)
	for _, conn := range conns {
		if _, ok := inactiveStatuses[conn.Status]; !ok {
			portPidMap[conn.Laddr] = conn
		}
	}

	ntrans := 0
	// remove outdated entries
	for a, ps := range ProcConnMap {
		if _, ok := portPidMap[a]; !ok && time.Since(ps.LastUpdate) > expiry {
			delete(ProcConnMap, a)
		} else if ok && ps.Pid == -1 {
			ntrans++
		}
	}

	// add new entries
	for a, c := range portPidMap {
		ps, ok := ProcConnMap[a]
		if !ok {
			ProcConnMap[a] = &ProcNetStat{Pid: c.Pid, NetCounters: IOCountersStat{}, RemoteAddr: c.Raddr, LastUpdate: time.Now()}
		} else if ps.Pid == -1 {
			ps.Pid = c.Pid
			ntrans--
		}
	}

	// Deal with remained transient connections
	if ntrans > 0 {
		guessPidByRemote(conns)
	}
}

// Here we are doing "best effort" when connection opened and closed between two polls.
// Pid is guessed by assuming that the application connects to the same remote endpoint repeatedly.
// This will be invalid if several applications connect to the same endpoint - hence the "guess".
func guessPidByRemote(conns []ConnectionStat) {
	portPidMap := make(map[Addr]ConnectionStat)
	for _, conn := range conns {
		portPidMap[conn.Raddr] = conn
	}
	for _, ps := range ProcConnMap {
		if ps.Pid == -1 {
			if ar, ok := portPidMap[ps.RemoteAddr]; ok {
				ps.Pid = ar.Pid
			}
		}
	}
}
