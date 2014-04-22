// +build freebsd

package gopsutil

func Virtual_memory() (Virtual_memoryStat, error) {
	ret := Virtual_memoryStat{}

	return ret, nil
}

func Swap_memory() (Swap_memoryStat, error) {
	ret := Swap_memoryStat{}

	return ret, nil
}
