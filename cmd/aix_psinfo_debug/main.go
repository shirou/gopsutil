package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: aix_psinfo_debug <pid>")
		return
	}

	pid := os.Args[1]
	psPath := fmt.Sprintf("/proc/%s/psinfo", pid)

	file, err := os.Open(psPath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", psPath, err)
		return
	}
	defer file.Close()

	var ps process.AIXPSInfo
	err = binary.Read(file, binary.BigEndian, &ps)
	if err != nil {
		fmt.Printf("Error reading struct: %v\n", err)
		return
	}

	fmt.Printf("Struct size: %d bytes\n", unsafe.Sizeof(ps))
	fmt.Printf("Struct from PID %s:\n", pid)
	fmt.Printf("  Flag: 0x%x\n", ps.Flag)
	fmt.Printf("  Flag2: 0x%x\n", ps.Flag2)
	fmt.Printf("  Nlwp: %d\n", ps.Nlwp)
	fmt.Printf("  Uid: %d\n", ps.Uid)
	fmt.Printf("  Euid: %d\n", ps.Euid)
	fmt.Printf("  Gid: %d\n", ps.Gid)
	fmt.Printf("  Egid: %d\n", ps.Egid)
	fmt.Printf("  Pid: %d\n", ps.Pid)
	fmt.Printf("  Ppid: %d\n", ps.Ppid)
	fmt.Printf("  Size: %d KB\n", ps.Size)
	fmt.Printf("  Rssize: %d KB\n", ps.Rssize)
	fmt.Printf("  Argc: %d\n", ps.Argc)
	fmt.Printf("  Argv: 0x%x\n", ps.Argv)
	fmt.Printf("  Envp: 0x%x\n", ps.Envp)

	fname := string(ps.Fname[:])
	fmt.Printf("  Fname (raw): %v\n", ps.Fname[:])
	fmt.Printf("  Fname (as string): '%s'\n", fname)

	// Trim nulls
	fnameLen := 0
	for i := 0; i < len(ps.Fname); i++ {
		if ps.Fname[i] == 0 {
			break
		}
		fnameLen++
	}
	fmt.Printf("  Fname (trimmed): '%s'\n", ps.Fname[:fnameLen])

	psargs := string(ps.Psargs[:])
	fmt.Printf("  Psargs (raw first 20 bytes): %v\n", ps.Psargs[:20])
	fmt.Printf("  Psargs (as string first 30 chars): '%s'\n", psargs[:minInt(30, len(psargs))])

	// Also print raw bytes at offset 0xa8 by seeking to that position
	file.Seek(0xa8, 0)
	rawBytes := make([]byte, 16)
	file.Read(rawBytes)
	fmt.Printf("\nRaw bytes from file at offset 0xa8: %v\n", rawBytes)
	fmt.Printf("As string: '%s'\n", string(rawBytes))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
