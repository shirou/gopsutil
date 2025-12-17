package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

type AIXStat struct {
	Flag   uint32  // process flags from proc struct p_flag
	Flag2  uint32  // process flags from proc struct p_flag2
	Flags  uint32  // /proc flags
	Nlwp   uint32  // number of threads in the process
	Stat   byte    // process state from proc p_stat
	Dmodel byte    // data model for the process
	Pad1   [6]byte // reserved for future use
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: aix_status_debug <pid>")
		os.Exit(1)
	}

	pid := os.Args[1]
	statusPath := "/proc/" + pid + "/status"

	data, err := os.ReadFile(statusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", statusPath, err)
		os.Exit(1)
	}

	if len(data) < 18 {
		fmt.Fprintf(os.Stderr, "Status file too small: %d bytes\n", len(data))
		os.Exit(1)
	}

	stat := AIXStat{
		Flag:   binary.BigEndian.Uint32(data[0:4]),
		Flag2:  binary.BigEndian.Uint32(data[4:8]),
		Flags:  binary.BigEndian.Uint32(data[8:12]),
		Nlwp:   binary.BigEndian.Uint32(data[12:16]),
		Stat:   data[16],
		Dmodel: data[17],
	}

	fmt.Printf("File: %s\n", statusPath)
	fmt.Printf("File size: %d bytes\n\n", len(data))
	fmt.Printf("Flag: 0x%08x\n", stat.Flag)
	fmt.Printf("Flag2: 0x%08x\n", stat.Flag2)
	fmt.Printf("Flags: 0x%08x\n", stat.Flags)
	fmt.Printf("Nlwp: 0x%08x (%d)\n", stat.Nlwp, stat.Nlwp)
	fmt.Printf("Stat: 0x%02x (%d, '%c')\n", stat.Stat, stat.Stat, rune(stat.Stat))
	fmt.Printf("Dmodel: 0x%02x (%d)\n", stat.Dmodel, stat.Dmodel)

	fmt.Println("\nFirst 32 bytes (hex):")
	for i := 0; i < 32 && i < len(data); i++ {
		fmt.Printf("%02x ", data[i])
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}
