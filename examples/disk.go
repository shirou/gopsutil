package main

import (
	"fmt"

	"github.com/shirou/gopsutil/disk"
)

func main() {
	d, _ := disk.Usage("/")
	fmt.Println("Disk usage: ", d)
}
