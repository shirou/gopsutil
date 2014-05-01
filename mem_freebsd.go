// +build freebsd

package gopsutil

func VirtualMemory() (*VirtualMemoryStat, error) {
	ret := &VirtualMemoryStat{}

	return ret, nil
}

func SwapMemory() (*SwapMemoryStat, error) {
	ret := &SwapMemoryStat{}

	return ret, nil
}
