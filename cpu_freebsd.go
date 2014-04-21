// +build freebsd

package main

import (
	"fmt"
)

func (c CPU) Cpu_times() ([]CPU_Times, error) {
	ret := make([]CPU_Times, 0)

	fmt.Println("FreeBSD")
	return ret, nil
}
