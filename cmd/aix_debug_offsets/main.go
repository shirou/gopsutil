// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	pid := os.Getpid()
	infoPath := fmt.Sprintf("/proc/%d/psinfo", pid)

	data, err := os.ReadFile(infoPath)
	if err != nil {
		fmt.Printf("ERROR reading %s: %v\n", infoPath, err)
		return
	}

	fmt.Printf("=== AIX psinfo Structure Analysis ===\nFile size: %d bytes\n\n", len(data))

	// Based on hex dump analysis:
	// The cmdline string is clearly visible at 0xa8 (168)
	// This should be Psargs[80]

	// Working backwards from 0xa8:
	// Psargs: 0xa8 (offset 168) - 80 bytes
	// Fname:  0xa0 (offset 160) - 16 bytes
	// Envp:   0x98 (offset 152) - 8 bytes
	// Argv:   0x90 (offset 144) - 8 bytes

	psargsStart := 168 // 0xa8
	fnameStart := 160  // 0xa0
	argvStart := 144   // 0x90
	envpStart := 152   // 0x98

	// Extract key fields using direct byte offsets
	psargs := make([]byte, 0)
	for i := psargsStart; i < psargsStart+80 && i < len(data); i++ {
		if data[i] == 0 {
			break
		}
		psargs = append(psargs, data[i])
	}

	// Fname is 16 bytes
	fname := data[fnameStart : fnameStart+16]
	fnameStr := extractString(fname)

	// Extract Argv and Envp (user-space pointers, uint64)
	argv := uint64(0)
	if argvStart+8 <= len(data) {
		argv = binary.BigEndian.Uint64(data[argvStart : argvStart+8])
	}
	envp := uint64(0)
	if envpStart+8 <= len(data) {
		envp = binary.BigEndian.Uint64(data[envpStart : envpStart+8])
	}

	fmt.Printf("Psargs (offset 0x%02x): %q\n", psargsStart, string(psargs))
	fmt.Printf("Fname  (offset 0x%02x): %q\n", fnameStart, fnameStr)
	fmt.Printf("Argv   (offset 0x%02x): 0x%016x\n", argvStart, argv)
	fmt.Printf("Envp   (offset 0x%02x): 0x%016x\n", envpStart, envp)

	// Try to find PID - should be a recognizable value
	fmt.Println("\n=== Scanning for PID ===")
	pidVal := int32(pid)
	pidBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(pidBytes, uint32(pidVal))

	for i := 0; i < len(data)-3; i++ {
		if data[i] == pidBytes[0] && data[i+1] == pidBytes[1] &&
			data[i+2] == pidBytes[2] && data[i+3] == pidBytes[3] {
			fmt.Printf("Found PID at offset 0x%02x\n", i)
		}
	}
}

func extractString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
		if c < 32 || c > 126 {
			// Non-printable, skip
			continue
		}
		// Return from first printable character
		return string(b[i:])
	}
	return ""
}
