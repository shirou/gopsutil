package load

import (
	"encoding/json"

	"github.com/shirou/gopsutil/internal/common"
)

var invoke common.Invoker

func init() {
	invoke = common.Invoke{}
}

type AvgStat struct {
	Load1  float64 `json:"load1" bson:"load1"`
	Load5  float64 `json:"load5" bson:"load5"`
	Load15 float64 `json:"load15" bson:"load15"`
}

func (l AvgStat) String() string {
	s, _ := json.Marshal(l)
	return string(s)
}

type MiscStat struct {
	ProcsRunning int `json:"procsRunning" bson:"procsRunning"`
	ProcsBlocked int `json:"procsBlocked" bson:"procsBlocked"`
	Ctxt         int `json:"ctxt" bson:"ctxt"`
}

func (m MiscStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
