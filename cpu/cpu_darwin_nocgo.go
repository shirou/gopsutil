// +build darwin
// +build !cgo

package cpu

import "github.com/shirou/gopsutil/internal/common"

func perCPUTimes() ([]TimesStat, error) {
	return []TimesStat{}, common.NotImplementedError
}

func allCPUTimes() ([]TimesStat, error) {
	return []TimesStat{}, common.NotImplementedError
}
