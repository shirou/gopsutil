package gopsutil

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

func test_getProcess() Process {
	checkPid := os.Getpid()
	if runtime.GOOS == "windows" {
		checkPid = 7960
	}
	ret, _ := NewProcess(int32(checkPid))
	return *ret
}

func Test_Pids(t *testing.T) {
	ret, err := Pids()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) == 0 {
		t.Errorf("could not get pids %v", ret)
	}
}

func Test_Pid_exists(t *testing.T) {
	checkPid := 1
	if runtime.GOOS == "windows" {
		checkPid = 0
	}

	ret, err := PidExists(int32(checkPid))
	if err != nil {
		t.Errorf("error %v", err)
	}

	if ret == false {
		t.Errorf("could not get init process %v", ret)
	}
}

func Test_NewProcess(t *testing.T) {
	checkPid := 1
	if runtime.GOOS == "windows" {
		checkPid = 0
	}

	ret, err := NewProcess(int32(checkPid))
	if err != nil {
		t.Errorf("error %v", err)
	}

	fmt.Println(ret)
}

func Test_Process_memory_maps(t *testing.T) {
	checkPid := os.Getpid()
	if runtime.GOOS == "windows" {
		checkPid = 0
	}
	return
	ret, err := NewProcess(int32(checkPid))

	mmaps, err := ret.MemoryMaps(false)
	if err != nil {
		t.Errorf("memory map get error %v", err)
	}
	for _, m := range *mmaps {
		fmt.Println(m)
	}

}

func Test_Process_Ppid(t *testing.T) {
	p := test_getProcess()

	v, err := p.Ppid()
	if err != nil {
		t.Errorf("geting ppid error %v", err)
	}
	if v == 0 {
		t.Errorf("return value is 0 %v", v)
	}

}

func Test_Process_IOCounters(t *testing.T) {
	p := test_getProcess()

	v, err := p.IOCounters()
	if err != nil {
		t.Errorf("geting ppid error %v", err)
		return
	}
	if v.ReadCount == 0 {
		t.Errorf("return value is 0 %v", v)
	}
	fmt.Println(v)

}

func Test_Process_NumCtx(t *testing.T) {
	p := test_getProcess()

	_, err := p.NumCtxSwitches()
	if err != nil {
		t.Errorf("geting numctx error %v", err)
		return
	}
}
