package process

import (
	"os"
	"runtime"
	"testing"
	"time"
)

func testGetProcess() Process {
	checkPid := os.Getpid() // process.test
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
	checkPid := os.Getpid()

	ret, err := PidExists(int32(checkPid))
	if err != nil {
		t.Errorf("error %v", err)
	}

	if ret == false {
		t.Errorf("could not get process exists %v", ret)
	}
}

func Test_NewProcess(t *testing.T) {
	checkPid := os.Getpid()
	if runtime.GOOS == "windows" {
		checkPid = 0
	}
	ret, err := NewProcess(int32(checkPid))
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &Process{}
	if runtime.GOOS != "windows" { // Windows pid is 0
		if empty == ret {
			t.Errorf("error %v", ret)
		}
	}

}

func Test_Process_memory_maps(t *testing.T) {
	checkPid := os.Getpid()
	if runtime.GOOS == "windows" {
		checkPid = 0
	}
	ret, err := NewProcess(int32(checkPid))

	mmaps, err := ret.MemoryMaps(false)
	if err != nil {
		t.Errorf("memory map get error %v", err)
	}
	empty := MemoryMapsStat{}
	for _, m := range *mmaps {
		if m == empty {
			t.Errorf("memory map get error %v", m)
		}
	}

}

func Test_Process_Ppid(t *testing.T) {
	p := testGetProcess()

	v, err := p.Ppid()
	if err != nil {
		t.Errorf("geting ppid error %v", err)
	}
	if v == 0 {
		t.Errorf("return value is 0 %v", v)
	}
}

func Test_Process_IOCounters(t *testing.T) {
	p := testGetProcess()

	v, err := p.IOCounters()
	if err != nil {
		t.Errorf("geting ppid error %v", err)
		return
	}
	empty := &IOCountersStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
}

func Test_Process_NumCtx(t *testing.T) {
	p := testGetProcess()

	_, err := p.NumCtxSwitches()
	if err != nil {
		t.Errorf("geting numctx error %v", err)
		return
	}
}

func Test_Process_Nice(t *testing.T) {
	p := testGetProcess()

	n, err := p.Nice()
	if err != nil {
		t.Errorf("geting nice error %v", err)
	}
	if n != 0 {
		t.Errorf("invalid nice: %d", n)
	}
}

func Test_Process_Name(t *testing.T) {
	p := testGetProcess()

	n, err := p.Name()
	if err != nil {
		t.Errorf("geting name error %v", err)
	}
	if n != "process.test" {
		t.Errorf("invalid name %s", n)
	}
}

func Test_Process_CpuPercent(t *testing.T) {
	p := testGetProcess()
	percent, err := p.CPUPercent(0)
	if err != nil {
		t.Errorf("error %v", err)
	}
	duration := time.Duration(1000) * time.Microsecond
	time.Sleep(duration)
	percent, err = p.CPUPercent(0)
	if err != nil {
		t.Errorf("error %v", err)
	}

	numcpu := runtime.NumCPU()
	if percent < 0.0 || percent > 100.0*float64(numcpu) {
		t.Fatalf("CPUPercent value is invalid: %f", percent)
	}
}

func Test_Process_CpuPercentLoop(t *testing.T) {
	p := testGetProcess()
	numcpu := runtime.NumCPU()

	for i := 0; i < 2; i++ {
		duration := time.Duration(100) * time.Microsecond
		percent, err := p.CPUPercent(duration)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if percent < 0.0 || percent > 100.0*float64(numcpu) {
			t.Fatalf("CPUPercent value is invalid: %f", percent)
		}
	}
}
