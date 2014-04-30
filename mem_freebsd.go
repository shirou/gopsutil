// +build freebsd

package gopsutil

func VirtualMemory() (VirtualMemoryStat, error) {
	ret := Virtual_memoryStat{}

	return ret, nil
}

func SwapMemory() (SwapMemoryStat, error) {
	ret := Swap_memoryStat{}

	return ret, nil
}
