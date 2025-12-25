// SPDX-License-Identifier: BSD-3-Clause
package nfs

import (
	"encoding/json"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var invoke common.Invoker = common.Invoke{}

// RPCClientStat represents RPC client statistics
type RPCClientStat struct {
	Calls      uint64 `json:"calls"`
	BadCalls   uint64 `json:"badCalls"`
	BadXIDs    uint64 `json:"badXids"`
	Timeouts   uint64 `json:"timeouts"`
	NewCreds   uint64 `json:"newCreds"`
	BadVerfs   uint64 `json:"badVerfs"`
	Timers     uint64 `json:"timers"`
	NoMem      uint64 `json:"noMem"`
	CantConn   uint64 `json:"cantConn"`
	Interrupts uint64 `json:"interrupts"`
	Retrans    uint64 `json:"retrans"`
	CantSend   uint64 `json:"cantSend"`
}

// RPCServerStat represents RPC server statistics
type RPCServerStat struct {
	Calls     uint64 `json:"calls"`
	BadCalls  uint64 `json:"badCalls"`
	NullRecv  uint64 `json:"nullRecv"`
	BadLen    uint64 `json:"badLen"`
	XdrCall   uint64 `json:"xdrCall"`
	DupChecks uint64 `json:"dupChecks"`
	DupReqs   uint64 `json:"dupReqs"`
}

// NFSClientStat represents NFS client statistics
type NFSClientStat struct {
	Calls      uint64            `json:"calls"`
	BadCalls   uint64            `json:"badCalls"`
	ClGets     uint64            `json:"clGets"`
	ClTooMany  uint64            `json:"clTooMany"`
	Operations map[string]uint64 `json:"operations"` // Per-operation counts (read, write, lookup, etc.)
}

// NFSServerStat represents NFS server statistics
type NFSServerStat struct {
	Calls      uint64            `json:"calls"`
	BadCalls   uint64            `json:"badCalls"`
	PublicV2   uint64            `json:"publicV2"`
	PublicV3   uint64            `json:"publicV3"`
	Operations map[string]uint64 `json:"operations"` // Per-operation counts (read, write, lookup, etc.)
}

// StatsStat represents complete NFS statistics
type StatsStat struct {
	RPCClientStats RPCClientStat `json:"rpcClient"`
	RPCServerStats RPCServerStat `json:"rpcServer"`
	NFSClientStats NFSClientStat `json:"nfsClient"`
	NFSServerStats NFSServerStat `json:"nfsServer"`
}

func (s StatsStat) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}
