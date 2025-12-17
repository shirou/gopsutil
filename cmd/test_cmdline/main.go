//go:build aix

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pid>\n", os.Args[0])
		os.Exit(1)
	}

	pidStr := os.Args[1]
	pidVal, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PID: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	p, err := process.NewProcessWithContext(ctx, int32(pidVal))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating process: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Process Info for PID %s ===\n\n", pidStr)

	// Test Name
	if name, err := p.NameWithContext(ctx); err == nil {
		fmt.Printf("Name: %s\n", name)
	} else {
		fmt.Printf("Name Error: %v\n", err)
	}

	// Test Cmdline
	if cmdline, err := p.CmdlineWithContext(ctx); err == nil {
		fmt.Printf("Cmdline: %s\n", cmdline)
	} else {
		fmt.Printf("Cmdline Error: %v\n", err)
	}

	// Test CmdlineSlice
	if cmdlineSlice, err := p.CmdlineSliceWithContext(ctx); err == nil {
		fmt.Printf("CmdlineSlice: %v\n", cmdlineSlice)
		fmt.Printf("  Length: %d\n", len(cmdlineSlice))
		for i, arg := range cmdlineSlice {
			fmt.Printf("  [%d]: %s\n", i, arg)
		}
	} else {
		fmt.Printf("CmdlineSlice Error: %v\n", err)
	}

	// Test Exe
	if exe, err := p.ExeWithContext(ctx); err == nil {
		fmt.Printf("Exe: %s\n", exe)
	} else {
		fmt.Printf("Exe Error: %v\n", err)
	}
}
