package host

import (
	"encoding/json"

	"github.com/shirou/gopsutil/internal/common"
)

var (
	invoke         common.Invoker
	cachedBootTime = uint64(0)
)

func init() {
	invoke = common.Invoke{}
}

// A HostInfoStat describes the host status.
// This is not in the psutil but it useful.
type InfoStat struct {
	Hostname             string `json:"hostname" bson:"hostname"`
	Uptime               uint64 `json:"uptime" bson:"uptime"`
	BootTime             uint64 `json:"bootTime" bson:"bootTime"`
	Procs                uint64 `json:"procs" bson:"procs"`                     // number of processes
	OS                   string `json:"os" bson:"os"`                           // ex: freebsd, linux
	Platform             string `json:"platform" bson:"platform"`               // ex: ubuntu, linuxmint
	PlatformFamily       string `json:"platformFamily" bson:"platformFamily"`   // ex: debian, rhel
	PlatformVersion      string `json:"platformVersion" bson:"platformVersion"` // version of the complete OS
	KernelVersion        string `json:"kernelVersion" bson:"kernelVersion"`     // version of the OS kernel (if available)
	VirtualizationSystem string `json:"virtualizationSystem" bson:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole" bson:"virtualizationRole"` // guest or host
	HostID               string `json:"hostid" bson:"hostid"`                         // ex: uuid
}

type UserStat struct {
	User     string `json:"user" bson:"user"`
	Terminal string `json:"terminal" bson:"terminal"`
	Host     string `json:"host" bson:"host"`
	Started  int    `json:"started" bson:"started"`
}

func (h InfoStat) String() string {
	s, _ := json.Marshal(h)
	return string(s)
}

func (u UserStat) String() string {
	s, _ := json.Marshal(u)
	return string(s)
}
