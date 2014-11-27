// +build darwin

package gopsutil

import (
	common "github.com/shirou/gopsutil/common"
)

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {

	return nil, common.NotImplementedError
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	return nil, common.NotImplementedError
}
