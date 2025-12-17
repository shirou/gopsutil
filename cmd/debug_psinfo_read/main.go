//go:build aix

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"

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

	pid := os.Args[1]
	psinfo := fmt.Sprintf("/proc/%s/psinfo", pid)

	file, err := os.Open(psinfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", psinfo, err)
		os.Exit(1)
	}
	defer file.Close()

	// Get file size
	fileInfo, _ := file.Stat()
	fmt.Printf("Psinfo file size: %d bytes\n\n", fileInfo.Size())

	// Read raw bytes first
	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Struct size: %d bytes\n", unsafe.Sizeof(process.AIXPSInfo{}))
	fmt.Printf("Struct fields offsets:\n")
	var p process.AIXPSInfo
	ptr := unsafe.Pointer(&p)
	fmt.Printf("  Fname offset: %d (0x%x)\n", uintptr(unsafe.Pointer(&p.Fname))-uintptr(ptr), uintptr(unsafe.Pointer(&p.Fname))-uintptr(ptr))
	fmt.Printf("  Psargs offset: %d (0x%x)\n", uintptr(unsafe.Pointer(&p.Psargs))-uintptr(ptr), uintptr(unsafe.Pointer(&p.Psargs))-uintptr(ptr))

	fmt.Printf("\nRaw bytes at offset 160 (Fname location): ")
	for i := 160; i < 176; i++ {
		fmt.Printf("%02x ", data[i])
	}
	fmt.Printf("\n")
	fmt.Printf("As ASCII: %q\n", string(data[160:176]))

	fmt.Printf("\nRaw bytes at offset 176 (Psargs location): ")
	for i := 176; i < 200; i++ {
		if i < len(data) {
			fmt.Printf("%02x ", data[i])
		}
	}
	fmt.Printf("\n")
	fmt.Printf("As ASCII: %q\n", string(data[176:256]))

	// Now try unmarshaling
	file.Seek(0, 0)
	var aixPSinfo process.AIXPSInfo
	err = binary.Read(file, binary.BigEndian, &aixPSinfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\nAfter binary.Read unmarshaling:\n")
	fmt.Printf("Fname field (raw): %q\n", string(aixPSinfo.Fname[:]))
	fmt.Printf("Fname extracted (from [0:]): %q\n", extractString(aixPSinfo.Fname[:]))
	fmt.Printf("Fname extracted (from [4:]): %q\n", extractString(aixPSinfo.Fname[4:]))
	fmt.Printf("Fname extracted (from [8:]): %q\n", extractString(aixPSinfo.Fname[8:]))
	fmt.Printf("Psargs field (first 20 bytes): %q\n", string(aixPSinfo.Psargs[:20]))
	fmt.Printf("Psargs extracted (from [0:]): %q\n", extractString(aixPSinfo.Psargs[:]))
	fmt.Printf("Psargs extracted (from [8:]): %q\n", extractString(aixPSinfo.Psargs[8:]))
}
