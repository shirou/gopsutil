// SPDX-License-Identifier: BSD-3-Clause
package process

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var mu sync.Mutex

func skipIfNotImplementedErr(t *testing.T, err error) {
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
}

func testGetProcess() Process {
	checkPid := os.Getpid() // process.test
	ret, _ := NewProcess(int32(checkPid))
	return *ret
}

func TestPids(t *testing.T) {
	ret, err := Pids()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) == 0 {
		t.Errorf("could not get pids %v", ret)
	}
}

func TestPid_exists(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := PidExists(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if !ret {
		t.Errorf("could not get process exists: %v", ret)
	}
}

func TestNewProcess(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	skipIfNotImplementedErr(t, err)
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

func TestMemoryMaps(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	// ungrouped memory maps
	mmaps, err := ret.MemoryMaps(false)
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
	mmaps, err = ret.MemoryMaps(true)
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
}

func TestMemoryInfo(t *testing.T) {
	p := testGetProcess()

	v, err := p.MemoryInfo()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting memory info error %v", err)
	}
	empty := MemoryInfoStat{}
	if v == nil || *v == empty {
		t.Errorf("could not get memory info %v", v)
	}
}

func TestCmdLine(t *testing.T) {
	p := testGetProcess()

	v, err := p.Cmdline()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting cmdline error %v", err)
	}
	if !strings.Contains(v, "process.test") {
		t.Errorf("invalid cmd line %v", v)
	}
}

func TestCmdLineSlice(t *testing.T) {
	p := testGetProcess()

	v, err := p.CmdlineSlice()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting cmdline slice error %v", err)
	}
	if !reflect.DeepEqual(v, os.Args) {
		t.Errorf("returned cmdline slice not as expected:\nexp: %v\ngot: %v", os.Args, v)
	}
}

func TestPpid(t *testing.T) {
	p := testGetProcess()

	v, err := p.Ppid()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting ppid error %v", err)
	}
	if v == 0 {
		t.Errorf("return value is 0 %v", v)
	}
	expected := os.Getppid()
	if v != int32(expected) {
		t.Errorf("return value is %v, expected %v", v, expected)
	}
}

func TestStatus(t *testing.T) {
	p := testGetProcess()

	v, err := p.Status()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting status error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get state")
	}
	if v[0] != Running && v[0] != Sleep {
		t.Errorf("got wrong state, %v", v)
	}
}

func TestTerminal(t *testing.T) {
	p := testGetProcess()

	_, err := p.Terminal()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting terminal error %v", err)
	}
}

func TestIOCounters(t *testing.T) {
	p := testGetProcess()

	v, err := p.IOCounters()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting iocounter error %v", err)
		return
	}
	empty := &IOCountersStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
}

func TestNumCtx(t *testing.T) {
	p := testGetProcess()

	_, err := p.NumCtxSwitches()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting numctx error %v", err)
		return
	}
}

func TestNice(t *testing.T) {
	p := testGetProcess()

	// https://github.com/shirou/gopsutil/issues/1532
	if os.Getenv("CI") == "true" && runtime.GOOS == "darwin" {
		t.Skip("Skip CI")
	}

	n, err := p.Nice()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting nice error %v", err)
	}
	if runtime.GOOS != "windows" && n != 0 && n != 20 && n != 8 {
		t.Errorf("invalid nice: %d", n)
	}
}

func TestGroups(t *testing.T) {
	p := testGetProcess()

	v, err := p.Groups()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting groups error %v", err)
	}
	if len(v) == 0 {
		t.Skip("Groups is empty")
	}
	if v[0] < 0 {
		t.Errorf("invalid Groups: %v", v)
	}
}

func TestNumThread(t *testing.T) {
	p := testGetProcess()

	n, err := p.NumThreads()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting NumThread error %v", err)
	}
	if n < 0 {
		t.Errorf("invalid NumThread: %d", n)
	}
}

func TestThreads(t *testing.T) {
	p := testGetProcess()

	n, err := p.NumThreads()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting NumThread error %v", err)
	}
	if n < 0 {
		t.Errorf("invalid NumThread: %d", n)
	}

	ts, err := p.Threads()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting Threads error %v", err)
	}
	if len(ts) != int(n) {
		t.Errorf("unexpected number of threads: %v vs %v", len(ts), n)
	}
}

func TestName(t *testing.T) {
	p := testGetProcess()

	n, err := p.Name()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting name error %v", err)
	}
	if !strings.Contains(n, "process.test") {
		t.Errorf("invalid Name %s", n)
	}
}

// #nosec G204
func TestLong_Name_With_Spaces(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("unable to create temp dir %v", err)
	}
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "loooong name with spaces.go")
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

	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	n, err := p.Name()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting name error %v", err)
	}
	basename := filepath.Base(tmpfile.Name() + ".exe")
	if basename != n {
		t.Fatalf("%s != %s", basename, n)
	}
	cmd.Process.Kill()
}

// #nosec G204
func TestLong_Name(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
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

	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	n, err := p.Name()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting name error %v", err)
	}
	basename := filepath.Base(tmpfile.Name() + ".exe")
	if basename != n {
		t.Fatalf("%s != %s", basename, n)
	}
	cmd.Process.Kill()
}

func TestName_Against_Python(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("only applies to posix")
	}
	py3Path, err := exec.LookPath("python3")
	if err != nil {
		t.Skipf("python3 not found: %s", err)
	}
	if out, err := exec.Command(py3Path, "-c", "import psutil").CombinedOutput(); err != nil {
		t.Skipf("psutil not found for %s: %s", py3Path, out)
	}

	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("unable to create temp dir %v", err)
	}
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "looooooooooooooooooooong.py")
	tmpfile, err := os.Create(tmpfilepath)
	if err != nil {
		t.Fatalf("unable to create temp file %v", err)
	}
	tmpfilecontent := []byte("#!" + py3Path + "\nimport psutil, time\nprint(psutil.Process().name(), flush=True)\nwhile True:\n\ttime.sleep(1)")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	if err := tmpfile.Chmod(0o744); err != nil {
		t.Fatalf("unable to chmod u+x temp file %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unable to close temp file %v", err)
	}
	cmd := exec.Command(tmpfilepath)
	outPipe, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(outPipe)
	cmd.Start()
	defer cmd.Process.Kill()
	scanner.Scan()
	pyName := scanner.Text() // first line printed by py3 script, its name
	t.Logf("pyName %s", pyName)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting process error %v", err)
	}
	name, err := p.Name()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting name error %v", err)
	}
	if pyName != name {
		t.Fatalf("psutil and gopsutil process.Name() results differ: expected %s, got %s", pyName, name)
	}
}

func TestExe(t *testing.T) {
	p := testGetProcess()

	n, err := p.Exe()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting Exe error %v", err)
	}
	if !strings.Contains(n, "process.test") {
		t.Errorf("invalid Exe %s", n)
	}
}

func TestCpuPercent(t *testing.T) {
	p := testGetProcess()
	_, err := p.Percent(0)
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	duration := time.Duration(1000) * time.Microsecond
	time.Sleep(duration)
	percent, err := p.Percent(0)
	if err != nil {
		t.Errorf("error %v", err)
	}

	numcpu := runtime.NumCPU()
	//	if percent < 0.0 || percent > 100.0*float64(numcpu) { // TODO
	if percent < 0.0 {
		t.Fatalf("CPUPercent value is invalid: %f, %d", percent, numcpu)
	}
}

func TestCpuPercentLoop(t *testing.T) {
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

func TestCreateTime(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skip CI")
	}

	p := testGetProcess()

	c, err := p.CreateTime()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if c < 1420000000 {
		t.Errorf("process created time is wrong.")
	}

	gotElapsed := time.Since(time.Unix(int64(c/1000), 0))
	maxElapsed := time.Duration(20 * time.Second)

	if gotElapsed >= maxElapsed {
		t.Errorf("this process has not been running for %v", gotElapsed)
	}
}

func TestParent(t *testing.T) {
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

func TestConnections(t *testing.T) {
	p := testGetProcess()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0") // dynamically get a random open port from OS
	if err != nil {
		t.Fatalf("unable to resolve localhost: %v", err)
	}
	l, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		t.Fatalf("unable to listen on %v: %v", addr, err)
	}
	defer l.Close()

	tcpServerAddr := l.Addr().String()
	tcpServerAddrIP := strings.Split(tcpServerAddr, ":")[0]
	tcpServerAddrPort, err := strconv.ParseUint(strings.Split(tcpServerAddr, ":")[1], 10, 32)
	if err != nil {
		t.Fatalf("unable to parse tcpServerAddr port: %v", err)
	}

	serverEstablished := make(chan struct{})
	go func() { // TCP listening goroutine
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		serverEstablished <- struct{}{}
		_, err = io.ReadAll(conn)
		if err != nil {
			panic(err)
		}
	}()

	conn, err := net.Dial("tcp", tcpServerAddr)
	if err != nil {
		t.Fatalf("unable to dial %v: %v", tcpServerAddr, err)
	}
	defer conn.Close()

	// Rarely the call to net.Dial returns before the server connection is
	// established. Wait so that the test doesn't fail.
	<-serverEstablished

	c, err := p.Connections()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("error %v", err)
	}
	if len(c) == 0 {
		t.Fatal("no connections found")
	}

	serverConnections := 0
	for _, connection := range c {
		if connection.Laddr.IP == tcpServerAddrIP && connection.Laddr.Port == uint32(tcpServerAddrPort) && connection.Raddr.Port != 0 {
			if connection.Status != "ESTABLISHED" {
				t.Fatalf("expected server connection to be ESTABLISHED, have %+v", connection)
			}
			serverConnections++
		}
	}

	clientConnections := 0
	for _, connection := range c {
		if connection.Raddr.IP == tcpServerAddrIP && connection.Raddr.Port == uint32(tcpServerAddrPort) {
			if connection.Status != "ESTABLISHED" {
				t.Fatalf("expected client connection to be ESTABLISHED, have %+v", connection)
			}
			clientConnections++
		}
	}

	if serverConnections != 1 { // two established connections, one for the server, the other for the client
		t.Fatalf("expected 1 server connection, have %d.\nDetails: %+v", serverConnections, c)
	}

	if clientConnections != 1 { // two established connections, one for the server, the other for the client
		t.Fatalf("expected 1 server connection, have %d.\nDetails: %+v", clientConnections, c)
	}
}

func TestChildren(t *testing.T) {
	p := testGetProcess()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "4")
	} else {
		cmd = exec.Command("sleep", "3")
	}
	require.NoError(t, cmd.Start())
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

func TestUsername(t *testing.T) {
	myPid := os.Getpid()
	currentUser, _ := user.Current()
	myUsername := currentUser.Username

	process, _ := NewProcess(int32(myPid))
	pidUsername, err := process.Username()
	skipIfNotImplementedErr(t, err)
	assert.Equal(t, myUsername, pidUsername)

	t.Log(pidUsername)
}

func TestCPUTimes(t *testing.T) {
	pid := os.Getpid()
	process, err := NewProcess(int32(pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	spinSeconds := 0.2
	cpuTimes0, err := process.Times()
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	// Spin for a duration of spinSeconds
	t0 := time.Now()
	tGoal := t0.Add(time.Duration(spinSeconds*1000) * time.Millisecond)
	require.NoError(t, err)
	for time.Now().Before(tGoal) {
		// This block intentionally left blank
	}

	cpuTimes1, err := process.Times()
	require.NoError(t, err)

	if cpuTimes0 == nil || cpuTimes1 == nil {
		t.FailNow()
	}
	measuredElapsed := cpuTimes1.Total() - cpuTimes0.Total()
	message := fmt.Sprintf("Measured %fs != spun time of %fs\ncpuTimes0=%v\ncpuTimes1=%v",
		measuredElapsed, spinSeconds, cpuTimes0, cpuTimes1)
	assert.Greater(t, measuredElapsed, float64(spinSeconds)/5, message)
	assert.Less(t, measuredElapsed, float64(spinSeconds)*5, message)
}

func TestOpenFiles(t *testing.T) {
	fp, err := os.Open("process_test.go")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, fp.Close())
	}()

	pid := os.Getpid()
	p, err := NewProcess(int32(pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	v, err := p.OpenFiles()
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmpty(t, v) // test always open files.

	for _, vv := range v {
		assert.NotEqual(t, "", vv.Path)
	}
}

func TestKill(t *testing.T) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "4")
	} else {
		cmd = exec.Command("sleep", "3")
	}
	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	err = p.Kill()
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	cmd.Wait()
}

func TestIsRunning(t *testing.T) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "localhost", "-n", "2")
	} else {
		cmd = exec.Command("sleep", "1")
	}
	cmd.Start()
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)
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

// #nosec G204
func TestEnviron(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf("unable to create temp dir %v", err)
	}
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "test.go")
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

	cmd.Env = []string{"testkey=envvalue"}

	require.NoError(t, cmd.Start())
	defer cmd.Process.Kill()
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	skipIfNotImplementedErr(t, err)
	require.NoError(t, err)

	envs, err := p.Environ()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("getting environ error %v", err)
	}
	var envvarFound bool
	for _, envvar := range envs {
		if envvar == "testkey=envvalue" {
			envvarFound = true
			break
		}
	}
	if !envvarFound {
		t.Error("environment variable not found")
	}
}

func TestCwd(t *testing.T) {
	myPid := os.Getpid()
	currentWorkingDirectory, _ := os.Getwd()

	process, _ := NewProcess(int32(myPid))
	pidCwd, err := process.Cwd()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Fatalf("getting cwd error %v", err)
	}
	pidCwd = strings.TrimSuffix(pidCwd, string(os.PathSeparator))
	assert.Equal(t, currentWorkingDirectory, pidCwd)

	t.Log(pidCwd)
}

func BenchmarkNewProcess(b *testing.B) {
	checkPid := os.Getpid()
	for i := 0; i < b.N; i++ {
		NewProcess(int32(checkPid))
	}
}

func BenchmarkProcessName(b *testing.B) {
	p := testGetProcess()
	for i := 0; i < b.N; i++ {
		p.Name()
	}
}

func BenchmarkProcessPpid(b *testing.B) {
	p := testGetProcess()
	for i := 0; i < b.N; i++ {
		p.Ppid()
	}
}

func BenchmarkProcesses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ps, err := Processes()
		require.NoError(b, err)
		require.NotEmpty(b, ps)
	}
}
