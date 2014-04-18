// +build windows

package main

import (
	"fmt"
	"syscall"
)

type Proc struct {
	Pid uint64
}

func GetProcList() ([]Proc, error) {
	ret := make([]Proc, 0)
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return ret, err
	}
	defer syscall.FreeLibrary(kernel32)
	Process32First, _ := syscall.GetProcAddress(kernel32, "Process32First")

	//	pFirst, _, err := syscall.Syscall(uintptr(Process32First), 0, 0, 0, 0)
	pFirst, _, err := syscall.Syscall(uintptr(Process32First), 0, 0, 0, 0)
	if err != nil {
		return ret, err
	}
	fmt.Printf("Proc: %v\n", pFirst)
	return ret, nil

}
