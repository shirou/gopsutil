package main

import (
	"fmt"

	"github.com/shirou/gopsutil/process"
)

func main() {
	processes, _ := process.Processes()
	fmt.Println("All processes: ", processes)
	for _, i := range processes {
		fmt.Println(i.Name())
	}
}
