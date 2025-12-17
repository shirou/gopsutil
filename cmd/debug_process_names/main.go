//go:build aix

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func extractString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return strings.TrimRight(string(b), "\x00")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pid>\n", os.Args[0])
		os.Exit(1)
	}

	pidStr := os.Args[1]

	// Read psinfo
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

	// Read status
	statusPath := fmt.Sprintf("/proc/%s/status", pidStr)
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading status: %v\n", err)
		os.Exit(1)
	}

	var aixStat process.AIXStat
	_ = aixStat // unused for now
	if len(statusData) >= 440 {
		// This is a simplified read - in reality we'd need to parse the binary structure properly
		fmt.Printf("Status file size: %d\n", len(statusData))
	}

	fmt.Printf("=== Process Info for PID %s ===\n", pidStr)
	fmt.Printf("Psargs (trimmed): %s\n", extractString(aixPSinfo.Psargs[:]))
	fmt.Printf("Fname (raw bytes): %v\n", aixPSinfo.Fname[:])
	fmt.Printf("Fname (trimmed): %s\n", extractString(aixPSinfo.Fname[:]))
	fmt.Printf("Argc: %d\n", aixPSinfo.Argc)
}
