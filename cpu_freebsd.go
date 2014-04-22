// +build freebsd

package gopsutil

func Cpu_times() ([]CPU_TimesStat, error) {
	ret := make([]CPU_TimesStat, 0)

	return ret, nil
}
