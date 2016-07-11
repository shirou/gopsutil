// +build linux

package common

import "os"

func NumProcs() (uint64, error) {
	f, err := os.Open(HostProc())
	if err != nil {
		return 0, err
	}

	list, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		return 0, err
	}
	return uint64(len(list)), err
}
