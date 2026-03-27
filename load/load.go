// SPDX-License-Identifier: BSD-3-Clause
package load

import (
	"encoding/json"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var invoke common.Invoker = common.Invoke{}

type AvgStat struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

func (l AvgStat) String() string {
	s, _ := json.Marshal(l)
	return string(s)
}

type MiscStat struct {
	ProcsTotal   int `json:"procsTotal"`
	ProcsCreated int `json:"procsCreated"`
	ProcsRunning int `json:"procsRunning"`
	ProcsBlocked int `json:"procsBlocked"`
	// Ctxt is the cumulative number of context switches since boot.
	// Populated on Linux (from /proc/stat) and AIX (from vmstat -s or perfstat).
	Ctxt int `json:"ctxt"`
	// SysCalls is the cumulative number of system calls since boot.
	// Not all platforms populate this field; zero means unavailable.
	// Populated on AIX (from vmstat -s or perfstat).
	SysCalls int `json:"sysCalls"`
	// Interrupts is the cumulative number of device interrupts since boot.
	// Not all platforms populate this field; zero means unavailable.
	// Populated on AIX (from vmstat -s or perfstat).
	Interrupts int `json:"interrupts"`
}

func (m MiscStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
