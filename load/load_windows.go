// +build windows

package gopsutil

import (
	common "github.com/shirou/gopsutil/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	ret := LoadAvgStat{}

	return &ret, common.NotImplementedError
}
