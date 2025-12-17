//go:build aix

package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pid>\n", os.Args[0])
		os.Exit(1)
	}

	pidStr := os.Args[1]
	psinfo := fmt.Sprintf("/proc/%s/psinfo", pidStr)

	file, err := os.Open(psinfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", psinfo, err)
		os.Exit(1)
	}
	defer file.Close()

	var aixPSinfo process.AIXPSInfo
	err = binary.Read(file, binary.BigEndian, &aixPSinfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading psinfo: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== AIXPSInfo for PID %s ===\n", pidStr)
	fmt.Printf("Argc: %d\n", aixPSinfo.Argc)
	fmt.Printf("Argv address: 0x%x (%d)\n", aixPSinfo.Argv, aixPSinfo.Argv)
	fmt.Printf("Envp address: 0x%x (%d)\n", aixPSinfo.Envp, aixPSinfo.Envp)
	fmt.Printf("Psargs: %s\n", string(aixPSinfo.Psargs[:]))
	fmt.Printf("Flag: 0x%x\n", aixPSinfo.Flag)
	fmt.Printf("Flag2: 0x%x\n", aixPSinfo.Flag2)
}
