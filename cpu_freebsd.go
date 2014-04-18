// +build freebsd

package main

import (
	"fmt"
)

func (c CPU) Cpu_times() map[string]string {
	ret := make(map[string]string)

	fmt.Println("FreeBSD")
	return ret
}
