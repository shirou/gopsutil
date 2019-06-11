package main

import (
	"fmt"

	"github.com/shirou/gopsutil/cpu"
)

func main() {
	info, _ := cpu.Info()
	fmt.Printf("Vendor ID: %s Model Name: %s Mhz: %f\n\n", info[0].VendorID, info[0].ModelName, info[0].Mhz)
	fmt.Println("All: ", info)
}
