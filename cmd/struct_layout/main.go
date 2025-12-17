package main

import (
	"fmt"
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	var ps process.AIXPSInfo

	fmt.Println("AIXPSInfo struct layout:")
	fmt.Printf("Total size: %d bytes\n\n", unsafe.Sizeof(ps))

	fmt.Printf("Flag:    offset %d, size %d\n", unsafe.Offsetof(ps.Flag), unsafe.Sizeof(ps.Flag))
	fmt.Printf("Flag2:   offset %d, size %d\n", unsafe.Offsetof(ps.Flag2), unsafe.Sizeof(ps.Flag2))
	fmt.Printf("Nlwp:    offset %d, size %d\n", unsafe.Offsetof(ps.Nlwp), unsafe.Sizeof(ps.Nlwp))
	fmt.Printf("Pad1:    offset %d, size %d\n", unsafe.Offsetof(ps.Pad1), unsafe.Sizeof(ps.Pad1))
	fmt.Printf("Uid:     offset %d, size %d\n", unsafe.Offsetof(ps.Uid), unsafe.Sizeof(ps.Uid))
	fmt.Printf("Euid:    offset %d, size %d\n", unsafe.Offsetof(ps.Euid), unsafe.Sizeof(ps.Euid))
	fmt.Printf("Gid:     offset %d, size %d\n", unsafe.Offsetof(ps.Gid), unsafe.Sizeof(ps.Gid))
	fmt.Printf("Egid:    offset %d, size %d\n", unsafe.Offsetof(ps.Egid), unsafe.Sizeof(ps.Egid))
	fmt.Printf("Pid:     offset %d, size %d\n", unsafe.Offsetof(ps.Pid), unsafe.Sizeof(ps.Pid))
	fmt.Printf("Ppid:    offset %d, size %d\n", unsafe.Offsetof(ps.Ppid), unsafe.Sizeof(ps.Ppid))
	fmt.Printf("Pgid:    offset %d, size %d\n", unsafe.Offsetof(ps.Pgid), unsafe.Sizeof(ps.Pgid))
	fmt.Printf("Sid:     offset %d, size %d\n", unsafe.Offsetof(ps.Sid), unsafe.Sizeof(ps.Sid))
	fmt.Printf("Ttydev:  offset %d, size %d\n", unsafe.Offsetof(ps.Ttydev), unsafe.Sizeof(ps.Ttydev))
	fmt.Printf("Addr:    offset %d, size %d\n", unsafe.Offsetof(ps.Addr), unsafe.Sizeof(ps.Addr))
	fmt.Printf("Size:    offset %d, size %d\n", unsafe.Offsetof(ps.Size), unsafe.Sizeof(ps.Size))
	fmt.Printf("Rssize:  offset %d, size %d\n", unsafe.Offsetof(ps.Rssize), unsafe.Sizeof(ps.Rssize))
	fmt.Printf("Start:   offset %d, size %d\n", unsafe.Offsetof(ps.Start), unsafe.Sizeof(ps.Start))
	fmt.Printf("Time:    offset %d, size %d\n", unsafe.Offsetof(ps.Time), unsafe.Sizeof(ps.Time))
	fmt.Printf("Cid:     offset %d, size %d\n", unsafe.Offsetof(ps.Cid), unsafe.Sizeof(ps.Cid))
	fmt.Printf("Pad2:    offset %d, size %d\n", unsafe.Offsetof(ps.Pad2), unsafe.Sizeof(ps.Pad2))
	fmt.Printf("Argc:    offset %d, size %d\n", unsafe.Offsetof(ps.Argc), unsafe.Sizeof(ps.Argc))
	fmt.Printf("Argv:    offset %d, size %d\n", unsafe.Offsetof(ps.Argv), unsafe.Sizeof(ps.Argv))
	fmt.Printf("Envp:    offset %d, size %d\n", unsafe.Offsetof(ps.Envp), unsafe.Sizeof(ps.Envp))
	fmt.Printf("Fname:   offset %d, size %d\n", unsafe.Offsetof(ps.Fname), unsafe.Sizeof(ps.Fname))
	fmt.Printf("Psargs:  offset %d, size %d\n", unsafe.Offsetof(ps.Psargs), unsafe.Sizeof(ps.Psargs))
	fmt.Printf("Pad:     offset %d, size %d\n", unsafe.Offsetof(ps.Pad), unsafe.Sizeof(ps.Pad))
	fmt.Printf("Lwp:     offset %d, size %d\n", unsafe.Offsetof(ps.Lwp), unsafe.Sizeof(ps.Lwp))
}
