package main

import (
	"fmt"

	"github.com/shirou/gopsutil/net"
)

func main() {
	io, _ := net.IOCounters(true)
	for j, i := range io {
		fmt.Printf("%d) %v \n", j, i)
	}
}
