// +build windows

package load

import (
	"github.com/shirou/gopsutil/internal/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	ret := LoadAvgStat{}

	return &ret, common.NotImplementedError
}

func Misc() (*MiscStat, error) {
	ret := MiscStat{}

	return &ret, common.NotImplementedError
}
