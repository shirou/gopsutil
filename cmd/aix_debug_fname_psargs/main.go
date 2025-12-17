// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// PrTimestruc64T from process_aix.go
type PrTimestruc64T struct {
	TvSec  int64  // 64 bit time_t value
	TvNsec int32  // 32 bit suseconds_t value
	Pad    uint32 // reserved for future use
}

type LwpsInfo struct {
	LwpId   uint64    // thread id
	Addr    uint64    // internal address of thread
	Wchan   uint64    // wait address for sleeping thread
	Flag    uint32    // thread flags
	Wtype   uint16    // type of thread wait
	State   byte      // thread state
	Sname   byte      // printable thread state character
	Nice    uint16    // nice value for CPU usage
	Pri     int32     // priority, high value = high priority
	Policy  uint32    // scheduling policy
	Clname  [8]byte   // printable scheduling policy string
	Onpro   int32     // processor on which thread last ran
	Bindpro int32     // processor to which thread is bound
	Ptid    uint32    // pthread id
	Pad1    uint32    // reserved for future use
	Pad     [7]uint64 // reserved for future use
}

// Simplified AIXPSInfo to test field alignment
type TestAIXPSInfo struct {
	Flag   uint32         // process flags
	Flag2  uint32         // process flags
	Nlwp   uint32         // number of threads
	Uid    uint32         // real user id
	Euid   uint32         // effective user id
	Gid    uint32         // real group id
	Egid   uint32         // effective group id
	Argc   uint32         // initial argument count
	Pid    uint64         // unique process id
	Ppid   uint64         // process id of parent
	Pgid   uint64         // pid of process group leader
	Sid    uint64         // session id
	Ttydev uint64         // controlling tty device
	Addr   uint64         // internal address
	Size   uint64         // size in KB
	Rssize uint64         // resident set size in KB
	Start  PrTimestruc64T // process start time
	Time   PrTimestruc64T // usr+sys cpu time
	Argv   uint64         // argv pointer
	Envp   uint64         // envp pointer
	Pad1   [16]uint64     // padding
	Fname  [16]byte       // executable name
	Psargs [80]byte       // argument list
}

func main() {
	pid := os.Getpid()
	infoPath := fmt.Sprintf("/proc/%d/psinfo", pid)

	data, err := os.ReadFile(infoPath)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("File size: %d bytes\n\n", len(data))

	// Test struct unmarshaling
	var psinfo TestAIXPSInfo
	err = binary.Read(os.NewFile(uintptr(0), infoPath), binary.BigEndian, &psinfo)
	if err == nil {
		fmt.Printf("Struct unmarshaling succeeded\n")
	} else {
		fmt.Printf("Struct size: %d bytes\n", binary.Size(psinfo))
	}

	// Read directly from file at specific offsets
	fmt.Println("\n=== Direct byte offset reads ===")

	// Fname should be at 0xa0 (160)
	if len(data) >= 176 {
		fnameBytes := data[160:176]
		fmt.Printf("Fname (0xa0): %q\n", extractString(fnameBytes))
	}

	// Psargs should be at 0xa8 (168)
	if len(data) >= 248 {
		psargsBytes := data[168:248]
		fmt.Printf("Psargs (0xa8): %q\n", extractString(psargsBytes))
	}

	// Try to unmarshal just the psinfo into a byte slice and inspect
	fmt.Println("\n=== Attempting struct unmarshal ===")
	file, _ := os.Open(infoPath)
	defer file.Close()
	err = binary.Read(file, binary.BigEndian, &psinfo)
	if err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
	} else {
		fmt.Printf("Unmarshal succeeded\n")
		fmt.Printf("Psinfo.Fname: %q\n", extractString(psinfo.Fname[:]))
		fmt.Printf("Psinfo.Psargs: %q\n", extractString(psinfo.Psargs[:]))
	}
}

func extractString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
