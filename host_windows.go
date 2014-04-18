// +build windows

package main

import (
	"os"
)

func (h Host) HostInfo() (HostInfo, error) {
	ret := HostInfo{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname

	return ret, nil
}
