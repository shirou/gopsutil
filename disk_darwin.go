// +build darwin

package gopsutil

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {

	return nil, NotImplementedError
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	return nil, NotImplementedError
}
