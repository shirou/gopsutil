package process

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shirou/gopsutil/internal/common"
	"github.com/stretchr/testify/assert"
)

var mu sync.Mutex

func skipIfNotImplementedErr(t *testing.T, err error) {
	if err == common.ErrNotImplementedError {
		t.Skip("not implemented")
	}
}

func testGetProcess() Process {
	checkPid := os.Getpid() // process.test
	ret, _ := NewProcess(int32(checkPid))
	return *ret
}

func testGetProcessWithFields(fields ...Field) Process {
	checkPid := os.Getpid() // process.test
	ret, _ := NewProcessWithFields(int32(checkPid), fields...)
	return *ret
}

func Test_Pids(t *testing.T) {
	ret, err := Pids()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) == 0 {
		t.Errorf("could not get pids %v", ret)
	}
}

func Test_Pids_Fail(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}

	mu.Lock()
	defer mu.Unlock()

	invoke = common.FakeInvoke{Suffix: "fail"}
	ret, err := Pids()
	skipIfNotImplementedErr(t, err)
	invoke = common.Invoke{}
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) != 9 {
		t.Errorf("wrong getted pid nums: %v/%d", ret, len(ret))
	}
}
func Test_Pid_exists(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := PidExists(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if ret == false {
		t.Errorf("could not get process exists: %v", ret)
	}
}

func Test_NewProcess(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	retWithFields, err := NewProcessWithFields(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	tests := []struct {
		name string
		proc *Process
	}{
		{
			name: "NewProcess",
			proc: ret,
		},
		{
			name: "NewProcessWithFields",
			proc: retWithFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			empty := &Process{}
			if runtime.GOOS != "windows" { // Windows pid is 0
				if empty == tt.proc {
					t.Errorf("error %v", tt.proc)
				}
			}
		})
	}

}

func Test_Process_memory_maps(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	retWithFields, err := NewProcessWithFields(int32(checkPid), FieldMemoryMaps, FieldMemoryMapsGrouped)
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	tests := []struct {
		name string
		proc *Process
	}{
		{
			name: "NewProcess",
			proc: ret,
		},
		{
			name: "NewProcessWithFields",
			proc: retWithFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ungrouped memory maps
			mmaps, err := tt.proc.MemoryMaps(false)
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("memory map get error %v", err)
			}
			empty := MemoryMapsStat{}
			for _, m := range *mmaps {
				if m == empty {
					t.Errorf("memory map get error %v", m)
				}
			}

			// grouped memory maps
			mmaps, err = tt.proc.MemoryMaps(true)
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("memory map get error %v", err)
			}
			if len(*mmaps) != 1 {
				t.Errorf("grouped memory maps length (%v) is not equal to 1", len(*mmaps))
			}
			if (*mmaps)[0] == empty {
				t.Errorf("memory map is empty")
			}
		})
	}
}
func Test_Process_MemoryInfo(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldMemoryInfo),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.MemoryInfo()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting memory info error %v", err)
			}
			empty := MemoryInfoStat{}
			if v == nil || *v == empty {
				t.Errorf("could not get memory info %v", v)
			}
		})
	}
}

func Test_Process_CmdLine(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldCmdline),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.Cmdline()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting cmdline error %v", err)
			}
			if !strings.Contains(v, "process.test") {
				t.Errorf("invalid cmd line %v", v)
			}
		})
	}
}

func Test_Process_CmdLineSlice(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldCmdlineSlice),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.CmdlineSlice()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Fatalf("getting cmdline slice error %v", err)
			}
			if !reflect.DeepEqual(v, os.Args) {
				t.Errorf("returned cmdline slice not as expected:\nexp: %v\ngot: %v", os.Args, v)
			}
		})
	}
}

func Test_Process_Ppid(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldPpid),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.Ppid()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting ppid error %v", err)
			}
			if v == 0 {
				t.Errorf("return value is 0 %v", v)
			}
		})
	}

}

func Test_Process_Status(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldStatus),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.Status()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting status error %v", err)
			}
			if v != "R" && v != "S" {
				t.Errorf("could not get state %v", v)
			}
		})
	}
}

func Test_Process_Terminal(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldTerminal),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.proc.Terminal()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting terminal error %v", err)
			}
		})
	}
}

func Test_Process_IOCounters(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldIOCounters),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.IOCounters()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting iocounter error %v", err)
				return
			}
			empty := &IOCountersStat{}
			if v == empty {
				t.Errorf("error %v", v)
			}
		})
	}
}

func Test_Process_NumCtx(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldNumCtxSwitches),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.proc.NumCtxSwitches()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting numctx error %v", err)
				return
			}
		})
	}
}

func Test_Process_Nice(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldNice),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.Nice()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting nice error %v", err)
			}
			if n != 0 && n != 20 && n != 8 {
				t.Errorf("invalid nice: %d", n)
			}
		})
	}
}
func Test_Process_NumThread(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldNumThreads),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.NumThreads()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting NumThread error %v", err)
			}
			if n < 0 {
				t.Errorf("invalid NumThread: %d", n)
			}
		})
	}
}

func Test_Process_Threads(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldNumThreads, FieldThreads),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.NumThreads()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting NumThread error %v", err)
			}
			if n < 0 {
				t.Errorf("invalid NumThread: %d", n)
			}

			ts, err := tt.proc.Threads()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting Threads error %v", err)
			}
			if len(ts) != int(n) {
				t.Errorf("unexpected number of threads: %v vs %v", len(ts), n)
			}
		})
	}
}

func Test_Process_Name(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.Name()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting name error %v", err)
			}
			if !strings.Contains(n, "process.test") {
				t.Errorf("invalid Exe %s", n)
			}
		})
	}
}
func Test_Process_Long_Name(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unable to create temp dir %v", err)
	}
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "looooooooooooooooooooong.go")
	tmpfile, err := os.Create(tmpfilepath)
	if err != nil {
		t.Fatalf("unable to create temp file %v", err)
	}

	tmpfilecontent := []byte("package main\nimport(\n\"time\"\n)\nfunc main(){\nfor range time.Tick(time.Second) {}\n}")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unable to close temp file %v", err)
	}

	err = exec.Command("go", "build", "-o", tmpfile.Name()+".exe", tmpfile.Name()).Run()
	if err != nil {
		t.Fatalf("unable to build temp file %v", err)
	}

	cmd := exec.Command(tmpfile.Name() + ".exe")

	assert.Nil(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)

	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)

	p2, err := NewProcessWithFields(int32(cmd.Process.Pid), FieldName)
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)

	tests := []struct {
		name string
		proc *Process
	}{
		{
			name: "NewProcess",
			proc: p,
		},
		{
			name: "NewProcessWithFields",
			proc: p2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.Name()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Fatalf("getting name error %v", err)
			}
			basename := filepath.Base(tmpfile.Name() + ".exe")
			if basename != n {
				t.Fatalf("%s != %s", basename, n)
			}
		})
	}

	cmd.Process.Kill()
}
func Test_Process_Exe(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldExe),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.proc.Exe()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("getting Exe error %v", err)
			}
			if !strings.Contains(n, "process.test") {
				t.Errorf("invalid Exe %s", n)
			}
		})
	}
}

func Test_Process_CpuPercent(t *testing.T) {
	p := testGetProcess()
	percent, err := p.Percent(0)
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	duration := time.Duration(1000) * time.Microsecond
	time.Sleep(duration)
	percent, err = p.Percent(0)
	if err != nil {
		t.Errorf("error %v", err)
	}

	numcpu := runtime.NumCPU()
	//	if percent < 0.0 || percent > 100.0*float64(numcpu) { // TODO
	if percent < 0.0 {
		t.Fatalf("CPUPercent value is invalid: %f, %d", percent, numcpu)
	}
}

func Test_Process_CpuPercentLoop(t *testing.T) {
	p := testGetProcess()
	numcpu := runtime.NumCPU()

	for i := 0; i < 2; i++ {
		duration := time.Duration(100) * time.Microsecond
		percent, err := p.Percent(duration)
		skipIfNotImplementedErr(t, err)
		if err != nil {
			t.Errorf("error %v", err)
		}
		//	if percent < 0.0 || percent > 100.0*float64(numcpu) { // TODO
		if percent < 0.0 {
			t.Fatalf("CPUPercent value is invalid: %f, %d", percent, numcpu)
		}
	}
}

func Test_Process_CreateTime(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" {
		t.Skip("Skip CI")
	}

	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldCreateTime),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := tt.proc.CreateTime()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}

			if c < 1420000000 {
				t.Errorf("process created time is wrong.")
			}

			gotElapsed := time.Since(time.Unix(int64(c/1000), 0))
			maxElapsed := time.Duration(5 * time.Second)

			if gotElapsed >= maxElapsed {
				t.Errorf("this process has not been running for %v", gotElapsed)
			}
		})
	}
}

func Test_Parent(t *testing.T) {
	p := testGetProcess()

	c, err := p.Parent()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("error %v", err)
	}
	if c == nil {
		t.Fatalf("could not get parent")
	}
	if c.Pid == 0 {
		t.Fatalf("wrong parent pid")
	}
}

func Test_Connections(t *testing.T) {
	ch0 := make(chan string)
	ch1 := make(chan string)
	go func() { // TCP listening goroutine
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0") // dynamically get a random open port from OS
		if err != nil {
			t.Skip("unable to resolve localhost:", err)
		}
		l, err := net.ListenTCP(addr.Network(), addr)
		if err != nil {
			t.Skip(fmt.Sprintf("unable to listen on %v: %v", addr, err))
		}
		defer l.Close()
		ch0 <- l.Addr().String()
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Skip("unable to accept connection:", err)
			}
			ch1 <- l.Addr().String()
			defer conn.Close()
		}
	}()
	go func() { // TCP client goroutine
		tcpServerAddr := <-ch0
		net.Dial("tcp", tcpServerAddr)
	}()

	tcpServerAddr := <-ch1
	tcpServerAddrIP := strings.Split(tcpServerAddr, ":")[0]
	tcpServerAddrPort, err := strconv.ParseUint(strings.Split(tcpServerAddr, ":")[1], 10, 32)
	if err != nil {
		t.Errorf("unable to parse tcpServerAddr port: %v", err)
	}

	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldConnections),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := tt.proc.Connections()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if len(c) == 0 {
				t.Errorf("no connections found")
			}
			found := 0
			for _, connection := range c {
				if connection.Status == "ESTABLISHED" && (connection.Laddr.IP == tcpServerAddrIP && connection.Laddr.Port == uint32(tcpServerAddrPort)) || (connection.Raddr.IP == tcpServerAddrIP && connection.Raddr.Port == uint32(tcpServerAddrPort)) {
					found++
				}
			}
			if found != 2 { // two established connections, one for the server, the other for the client
				t.Errorf(fmt.Sprintf("wrong connections: %+v", c))
			}
		})
	}
}

func Test_Children(t *testing.T) {
	p := testGetProcess()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "4")
	} else {
		cmd = exec.Command("sleep", "3")
	}
	assert.Nil(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)

	c, err := p.Children()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("error %v", err)
	}
	if len(c) == 0 {
		t.Fatalf("children is empty")
	}
	found := false
	for _, child := range c {
		if child.Pid == int32(cmd.Process.Pid) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("could not find child %d", cmd.Process.Pid)
	}
}

func Test_Username(t *testing.T) {
	currentUser, _ := user.Current()
	myUsername := currentUser.Username

	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldUsername),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pidUsername, err := tt.proc.Username()
			skipIfNotImplementedErr(t, err)
			assert.Equal(t, myUsername, pidUsername)
			t.Log(pidUsername)
		})
	}

}

func Test_CPUTimes(t *testing.T) {
	pid := os.Getpid()
	process, err := NewProcess(int32(pid))
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)

	spinSeconds := 0.2
	cpuTimes0, err := process.Times()
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)

	// Spin for a duration of spinSeconds
	t0 := time.Now()
	tGoal := t0.Add(time.Duration(spinSeconds*1000) * time.Millisecond)
	assert.Nil(t, err)
	for time.Now().Before(tGoal) {
		// This block intentionally left blank
	}

	cpuTimes1, err := process.Times()
	assert.Nil(t, err)

	if cpuTimes0 == nil || cpuTimes1 == nil {
		t.FailNow()
	}
	measuredElapsed := cpuTimes1.Total() - cpuTimes0.Total()
	message := fmt.Sprintf("Measured %fs != spun time of %fs\ncpuTimes0=%v\ncpuTimes1=%v",
		measuredElapsed, spinSeconds, cpuTimes0, cpuTimes1)
	assert.True(t, measuredElapsed > float64(spinSeconds)/5, message)
	assert.True(t, measuredElapsed < float64(spinSeconds)*5, message)
}

func Test_OpenFiles(t *testing.T) {
	tests := []struct {
		name string
		proc Process
	}{
		{
			name: "NewProcess",
			proc: testGetProcess(),
		},
		{
			name: "NewProcessWithFields",
			proc: testGetProcessWithFields(FieldOpenFiles),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := tt.proc.OpenFiles()
			skipIfNotImplementedErr(t, err)
			assert.Nil(t, err)
			assert.NotEmpty(t, v) // test always open files.

			for _, vv := range v {
				assert.NotEqual(t, "", vv.Path)
			}
		})
	}
}

func Test_Kill(t *testing.T) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "4")
	} else {
		cmd = exec.Command("sleep", "3")
	}
	assert.Nil(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)
	err = p.Kill()
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)
	cmd.Wait()
}

func Test_IsRunning(t *testing.T) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "2")
	} else {
		cmd = exec.Command("sleep", "1")
	}
	cmd.Start()
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	assert.Nil(t, err)
	running, err := p.IsRunning()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("IsRunning error: %v", err)
	}
	if !running {
		t.Fatalf("process should be found running")
	}
	cmd.Wait()
	running, err = p.IsRunning()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("IsRunning error: %v", err)
	}
	if running {
		t.Fatalf("process should NOT be found running")
	}
}

func Test_AllProcesses_cmdLine(t *testing.T) {
	procs, err := Processes()
	if err == nil {
		for _, proc := range procs {
			var exeName string
			var cmdLine string

			exeName, _ = proc.Exe()
			cmdLine, err = proc.Cmdline()
			if err != nil {
				cmdLine = "Error: " + err.Error()
			}

			t.Logf("Process #%v: Name: %v / CmdLine: %v\n", proc.Pid, exeName, cmdLine)
		}
	}
}

func Test_Processes(t *testing.T) {
	myPID := os.Getpid()

	procs, err := Processes()
	if err != nil {
		t.Fatal(err)
	}

	procsFields, err := ProcessesWithFields(context.Background(), FieldCmdlineSlice)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		procs []*Process
	}{
		{
			name:  "Processes",
			procs: procs,
		},
		{
			name:  "ProcessesWithFields",
			procs: procsFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false

			for _, proc := range tt.procs {
				if proc.Pid == int32(myPID) {
					found = true
					v, err := proc.CmdlineSlice()
					skipIfNotImplementedErr(t, err)
					if err != nil {
						t.Fatalf("getting cmdline slice error %v", err)
					}
					if !reflect.DeepEqual(v, os.Args) {
						t.Errorf("returned cmdline slice not as expected:\nexp: %v\ngot: %v", os.Args, v)
					}

					break
				}
			}

			if !found {
				t.Error("not found myself in Processes list")
			}
		})
	}
}

func Test_AllFields(t *testing.T) {
	procAllFields := testGetProcessWithFields(AllFields...)

	var (
		resultDynamic, resultField, resultAllFields     interface{}
		errDynamic, errField, errAllFields, errNoFields error
	)

	for _, f := range AllFields {
		procDynamic := testGetProcess()
		procField := testGetProcessWithFields(f)
		procNoField := testGetProcessWithFields()
		stableValue := true

		switch f {
		case FieldBackground:
			resultDynamic, errDynamic = procDynamic.Background()
			resultField, errField = procField.Background()
			resultAllFields, errAllFields = procAllFields.Background()
			_, errNoFields = procNoField.Background()
		case FieldCmdline:
			resultDynamic, errDynamic = procDynamic.Cmdline()
			resultField, errField = procField.Cmdline()
			resultAllFields, errAllFields = procAllFields.Cmdline()
			_, errNoFields = procNoField.Cmdline()
		case FieldCmdlineSlice:
			resultDynamic, errDynamic = procDynamic.CmdlineSlice()
			resultField, errField = procField.CmdlineSlice()
			resultAllFields, errAllFields = procAllFields.CmdlineSlice()
			_, errNoFields = procNoField.CmdlineSlice()
		case FieldConnections:
			resultDynamic, errDynamic = procDynamic.Connections()
			resultField, errField = procField.Connections()
			resultAllFields, errAllFields = procAllFields.Connections()
			_, errNoFields = procNoField.Connections()
		case FieldCPUPercent:
			resultDynamic, errDynamic = procDynamic.CPUPercent()
			resultField, errField = procField.CPUPercent()
			resultAllFields, errAllFields = procAllFields.CPUPercent()
			_, errNoFields = procNoField.CPUPercent()
			stableValue = false
		case FieldCreateTime:
			resultDynamic, errDynamic = procDynamic.CreateTime()
			resultField, errField = procField.CreateTime()
			resultAllFields, errAllFields = procAllFields.CreateTime()
			_, errNoFields = procNoField.CreateTime()
		case FieldCwd:
			resultDynamic, errDynamic = procDynamic.Cwd()
			resultField, errField = procField.Cwd()
			resultAllFields, errAllFields = procAllFields.Cwd()
			_, errNoFields = procNoField.Cwd()
		case FieldExe:
			resultDynamic, errDynamic = procDynamic.Exe()
			resultField, errField = procField.Exe()
			resultAllFields, errAllFields = procAllFields.Exe()
			_, errNoFields = procNoField.Exe()
		case FieldForeground:
			resultDynamic, errDynamic = procDynamic.Foreground()
			resultField, errField = procField.Foreground()
			resultAllFields, errAllFields = procAllFields.Foreground()
			_, errNoFields = procNoField.Foreground()
		case FieldGids:
			resultDynamic, errDynamic = procDynamic.Gids()
			resultField, errField = procField.Gids()
			resultAllFields, errAllFields = procAllFields.Gids()
			_, errNoFields = procNoField.Gids()
		case FieldIOCounters:
			resultDynamic, errDynamic = procDynamic.IOCounters()
			resultField, errField = procField.IOCounters()
			resultAllFields, errAllFields = procAllFields.IOCounters()
			_, errNoFields = procNoField.IOCounters()
			stableValue = false
		case FieldIOnice:
			resultDynamic, errDynamic = procDynamic.IOnice()
			resultField, errField = procField.IOnice()
			resultAllFields, errAllFields = procAllFields.IOnice()
			_, errNoFields = procNoField.IOnice()
		case FieldIsRunning:
			resultDynamic, errDynamic = procDynamic.IsRunning()
			resultField, errField = procField.IsRunning()
			resultAllFields, errAllFields = procAllFields.IsRunning()
			_, errNoFields = procNoField.IsRunning()
		case FieldMemoryInfo:
			resultDynamic, errDynamic = procDynamic.MemoryInfo()
			resultField, errField = procField.MemoryInfo()
			resultAllFields, errAllFields = procAllFields.MemoryInfo()
			_, errNoFields = procNoField.MemoryInfo()
			stableValue = false
		case FieldMemoryInfoEx:
			resultDynamic, errDynamic = procDynamic.MemoryInfoEx()
			resultField, errField = procField.MemoryInfoEx()
			resultAllFields, errAllFields = procAllFields.MemoryInfoEx()
			_, errNoFields = procNoField.MemoryInfoEx()
			stableValue = false
		case FieldMemoryMaps:
			resultDynamic, errDynamic = procDynamic.MemoryMaps(false)
			resultField, errField = procField.MemoryMaps(false)
			resultAllFields, errAllFields = procAllFields.MemoryMaps(false)
			_, errNoFields = procNoField.MemoryMaps(false)
			stableValue = false
		case FieldMemoryMapsGrouped:
			resultDynamic, errDynamic = procDynamic.MemoryMaps(true)
			resultField, errField = procField.MemoryMaps(true)
			resultAllFields, errAllFields = procAllFields.MemoryMaps(true)
			_, errNoFields = procNoField.MemoryMaps(true)
			stableValue = false
		case FieldMemoryPercent:
			resultDynamic, errDynamic = procDynamic.MemoryPercent()
			resultField, errField = procField.MemoryPercent()
			resultAllFields, errAllFields = procAllFields.MemoryPercent()
			_, errNoFields = procNoField.MemoryPercent()
			stableValue = false
		case FieldName:
			resultDynamic, errDynamic = procDynamic.Name()
			resultField, errField = procField.Name()
			resultAllFields, errAllFields = procAllFields.Name()
			_, errNoFields = procNoField.Name()
		case FieldNetIOCounters:
			resultDynamic, errDynamic = procDynamic.NetIOCounters(false)
			resultField, errField = procField.NetIOCounters(false)
			resultAllFields, errAllFields = procAllFields.NetIOCounters(false)
			_, errNoFields = procNoField.NetIOCounters(false)
			stableValue = false
		case FieldNetIOCountersPerNic:
			resultDynamic, errDynamic = procDynamic.NetIOCounters(true)
			resultField, errField = procField.NetIOCounters(true)
			resultAllFields, errAllFields = procAllFields.NetIOCounters(true)
			_, errNoFields = procNoField.NetIOCounters(true)
			stableValue = false
		case FieldNice:
			resultDynamic, errDynamic = procDynamic.Nice()
			resultField, errField = procField.Nice()
			resultAllFields, errAllFields = procAllFields.Nice()
			_, errNoFields = procNoField.Nice()
		case FieldNumCtxSwitches:
			resultDynamic, errDynamic = procDynamic.NumCtxSwitches()
			resultField, errField = procField.NumCtxSwitches()
			resultAllFields, errAllFields = procAllFields.NumCtxSwitches()
			_, errNoFields = procNoField.NumCtxSwitches()
			stableValue = false
		case FieldNumFDs:
			resultDynamic, errDynamic = procDynamic.NumFDs()
			resultField, errField = procField.NumFDs()
			resultAllFields, errAllFields = procAllFields.NumFDs()
			_, errNoFields = procNoField.NumFDs()
		case FieldNumThreads:
			resultDynamic, errDynamic = procDynamic.NumThreads()
			resultField, errField = procField.NumThreads()
			resultAllFields, errAllFields = procAllFields.NumThreads()
			_, errNoFields = procNoField.NumThreads()
		case FieldOpenFiles:
			resultDynamic, errDynamic = procDynamic.OpenFiles()
			resultField, errField = procField.OpenFiles()
			resultAllFields, errAllFields = procAllFields.OpenFiles()
			_, errNoFields = procNoField.OpenFiles()
		case FieldPageFaults:
			resultDynamic, errDynamic = procDynamic.PageFaults()
			resultField, errField = procField.PageFaults()
			resultAllFields, errAllFields = procAllFields.PageFaults()
			_, errNoFields = procNoField.PageFaults()
			stableValue = false
		case FieldPpid:
			resultDynamic, errDynamic = procDynamic.Ppid()
			resultField, errField = procField.Ppid()
			resultAllFields, errAllFields = procAllFields.Ppid()
			_, errNoFields = procNoField.Ppid()
		case FieldRlimit:
			resultDynamic, errDynamic = procDynamic.Rlimit()
			resultField, errField = procField.Rlimit()
			resultAllFields, errAllFields = procAllFields.Rlimit()
			_, errNoFields = procNoField.Rlimit()
		case FieldRlimitUsage:
			resultDynamic, errDynamic = procDynamic.RlimitUsage(true)
			resultField, errField = procField.RlimitUsage(true)
			resultAllFields, errAllFields = procAllFields.RlimitUsage(true)
			_, errNoFields = procNoField.RlimitUsage(true)
			stableValue = false
		case FieldStatus:
			resultDynamic, errDynamic = procDynamic.Status()
			resultField, errField = procField.Status()
			resultAllFields, errAllFields = procAllFields.Status()
			_, errNoFields = procNoField.Status()
		case FieldTerminal:
			resultDynamic, errDynamic = procDynamic.Terminal()
			resultField, errField = procField.Terminal()
			resultAllFields, errAllFields = procAllFields.Terminal()
			_, errNoFields = procNoField.Terminal()
		case FieldTgid:
			resultDynamic, errDynamic = procDynamic.Tgid()
			resultField, errField = procField.Tgid()
			resultAllFields, errAllFields = procAllFields.Tgid()
			_, errNoFields = procNoField.Tgid()
		case FieldThreads:
			resultDynamic, errDynamic = procDynamic.Threads()
			resultField, errField = procField.Threads()
			resultAllFields, errAllFields = procAllFields.Threads()
			_, errNoFields = procNoField.Threads()
			stableValue = false
		case FieldTimes:
			resultDynamic, errDynamic = procDynamic.Times()
			resultField, errField = procField.Times()
			resultAllFields, errAllFields = procAllFields.Times()
			_, errNoFields = procNoField.Times()
			stableValue = false
		case FieldUids:
			resultDynamic, errDynamic = procDynamic.Uids()
			resultField, errField = procField.Uids()
			resultAllFields, errAllFields = procAllFields.Uids()
			_, errNoFields = procNoField.Uids()
		case FieldUsername:
			resultDynamic, errDynamic = procDynamic.Username()
			resultField, errField = procField.Username()
			resultAllFields, errAllFields = procAllFields.Username()
			_, errNoFields = procNoField.Username()
		}

		if errNoFields != ErrorFieldNotRequested && errNoFields != common.ErrNotImplementedError {
			t.Errorf("Field %v: procNoField err = %v, want %v", f, errNoFields, ErrorFieldNotRequested)
		}

		if f.String() == "unknown" {
			t.Errorf("Field #%d don't have a String() value", f)
		}

		if stableValue && !reflect.DeepEqual(resultField, resultDynamic) {
			t.Errorf("procField.%v() = %v, want %v", f, resultField, resultDynamic)
		}
		if stableValue && !reflect.DeepEqual(resultAllFields, resultDynamic) {
			t.Errorf("procAllFields.%v() = %v, want %v", f, resultAllFields, resultDynamic)
		}
		if errField != errDynamic {
			t.Errorf("procField.%v() error = %v, want %v", f, errField, errDynamic)
		}
		if errAllFields != errDynamic {
			t.Errorf("procAllFields.%v() error = %v, want %v", f, errAllFields, errDynamic)
		}
	}
}

// Benchmark_NewProcess test that NewProcessWithFields provide performance gain.
func Benchmark_NewProcessWithFields(b *testing.B) {
	checkPid := int32(os.Getpid())

	for _, name := range []string{"NewProcess", "WithFields"} {
		b.Run(name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				var (
					proc *Process
					err  error
				)

				if name == "NewProcess" {
					proc, err = NewProcess(checkPid)
				} else {
					proc, err = NewProcessWithFields(checkPid, FieldUsername, FieldCPUPercent, FieldMemoryPercent, FieldStatus, FieldMemoryInfo, FieldTimes, FieldCmdline)
				}

				if err != nil {
					b.Error(err)
					return
				}

				_, err = proc.Username()
				if err != nil {
					b.Error(err)
					return
				}

				_, err = proc.CPUPercent()
				if err != nil {
					b.Error(err)
					return
				}

				_, err = proc.MemoryPercent()
				if err != nil {
					b.Error(err)
					return
				}
				_, err = proc.Status()
				if err != nil {
					b.Error(err)
					return
				}
				_, err = proc.MemoryInfo()
				if err != nil {
					b.Error(err)
					return
				}
				_, err = proc.Times()
				if err != nil {
					b.Error(err)
					return
				}
				_, err = proc.Cmdline()
				if err != nil {
					b.Error(err)
					return
				}
			}
		})
	}
}
