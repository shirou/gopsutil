// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package main

import (
	"encoding/hex"
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

	fmt.Printf("File: %s\n", infoPath)
	fmt.Printf("Size: %d bytes\n\n", len(data))

	// Print hex dump
	fmt.Println("HEX DUMP:")
	fmt.Println(hex.Dump(data[:min(512, len(data))]))

	// Print ASCII where possible
	fmt.Println("\nRAW BYTES (first 256):")
	for i := 0; i < min(256, len(data)); i++ {
		b := data[i]
		if b >= 32 && b < 127 {
			fmt.Printf("%c", b)
		} else {
			fmt.Printf("[%02x]", b)
		}
		if (i+1)%32 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
