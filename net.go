package gopsutil

import (
	"encoding/json"
)

type NetIOCountersStat struct {
	Name        string `json:"name"`         // interface name
	BytesSent   uint64 `json:"bytes_sent"`   // number of bytes sent
	BytesRecv   uint64 `json:"bytes_recv"`   // number of bytes received
	PacketsSent uint64 `json:"packets_sent"` // number of packets sent
	PacketsRecv uint64 `json:"packets_recv"` // number of packets received
	Errin       uint64 `json:"errin"`        // total number of errors while receiving
	Errout      uint64 `json:"errout"`       // total number of errors while sending
	Dropin      uint64 `json:"dropin"`       // total number of incoming packets which were dropped
	Dropout     uint64 `json:"dropout"`      // total number of outgoing packets which were dropped (always 0 on OSX and BSD)
}

type Addr struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
}

type NetConnectionStat struct {
	Fd     uint32 `json:"fd"`
	Family uint32 `json:"family"`
	Type   uint32 `json:"type"`
	Laddr  Addr   `json:"laddr"`
	Raddr  Addr   `json:"raddr"`
	Status string `json:"status"`
	Pid    int32  `json:"pid"`
}

func (n NetIOCountersStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func (n NetConnectionStat) String() string {
	s, _ := json.Marshal(n)
	return string(s)
}

func (a Addr) String() string {
	s, _ := json.Marshal(a)
	return string(s)
}
