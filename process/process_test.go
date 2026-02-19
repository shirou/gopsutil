// SPDX-License-Identifier: BSD-3-Clause
package process

import (
	"bufio"
	"context"
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
	"slices"
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

func testGetProcess() Process {
	checkPid := os.Getpid() // process.test
	ret, _ := NewProcess(int32(checkPid))
	return *ret
}

func TestPids(t *testing.T) {
	ret, err := Pids()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	assert.NotEmptyf(t, ret, "could not get pids %v", ret)
}

func TestPid_exists(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := PidExists(int32(checkPid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	assert.Truef(t, ret, "could not get process exists: %v", ret)
}

func TestNewProcess(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	empty := &Process{}
	if runtime.GOOS != "windows" { // Windows pid is 0
		assert.NotSamef(t, empty, ret, "error %v", ret)
	}
}

func TestMemoryMaps(t *testing.T) {
	checkPid := os.Getpid()

	ret, err := NewProcess(int32(checkPid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	// ungrouped memory maps
	mmaps, err := ret.MemoryMaps(false)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "memory map get error %v", err)
	empty := MemoryMapsStat{}
	for _, m := range *mmaps {
		assert.NotEqualf(t, m, empty, "memory map get error %v", m)
	}

	// grouped memory maps
	mmaps, err = ret.MemoryMaps(true)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "memory map get error %v", err)
	assert.Lenf(t, *mmaps, 1, "grouped memory maps length (%v) is not equal to 1", len(*mmaps))
	assert.NotEqualf(t, (*mmaps)[0], empty, "memory map is empty")
}

func TestMemoryInfo(t *testing.T) {
	p := testGetProcess()

	v, err := p.MemoryInfo()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting memory info error %v", err)
	empty := MemoryInfoStat{}
	if v == nil || *v == empty {
		t.Errorf("could not get memory info %v", v)
	}
}

func TestCmdLine(t *testing.T) {
	p := testGetProcess()

	v, err := p.Cmdline()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting cmdline error %v", err)
	assert.Containsf(t, v, "process.test", "invalid cmd line %v", v)
}

func TestCmdLineSlice(t *testing.T) {
	p := testGetProcess()

	v, err := p.CmdlineSlice()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting cmdline slice error %v", err)
	assert.Truef(t, reflect.DeepEqual(v, os.Args), "returned cmdline slice not as expected:\nexp: %v\ngot: %v", os.Args, v)
}

func TestPpid(t *testing.T) {
	p := testGetProcess()

	v, err := p.Ppid()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting ppid error %v", err)
	assert.NotZerof(t, v, "return value is 0 %v", v)
	expected := os.Getppid()
	assert.Equalf(t, int32(expected), v, "return value is %v, expected %v", v, expected)
}

func TestStatus(t *testing.T) {
	p := testGetProcess()

	v, err := p.Status()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting status error %v", err)
	assert.NotEmptyf(t, v, "could not get state")
	if v[0] != Running && v[0] != Sleep {
		t.Errorf("got wrong state, %v", v)
	}
}

func TestTerminal(t *testing.T) {
	p := testGetProcess()

	_, err := p.Terminal()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	assert.NoErrorf(t, err, "getting terminal error %v", err)
}

func TestIOCounters(t *testing.T) {
	p := testGetProcess()

	v, err := p.IOCounters()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting iocounter error %v", err)
	empty := &IOCountersStat{}
	assert.NotSamef(t, v, empty, "error %v", v)
}

func TestNumCtx(t *testing.T) {
	p := testGetProcess()

	_, err := p.NumCtxSwitches()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	assert.NoErrorf(t, err, "getting numctx error %v", err)
}

func TestNice(t *testing.T) {
	p := testGetProcess()

	// https://github.com/shirou/gopsutil/issues/1532
	if os.Getenv("CI") == "true" && runtime.GOOS == "darwin" {
		t.Skip("Skip CI")
	}

	n, err := p.Nice()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting nice error %v", err)
	if runtime.GOOS != "windows" && n != 0 && n != 20 && n != 8 {
		t.Errorf("invalid nice: %d", n)
	}
}

func TestGroups(t *testing.T) {
	p := testGetProcess()

	v, err := p.Groups()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting groups error %v", err)
	if len(v) == 0 {
		t.Skip("Groups is empty")
	}
}

func TestNumThread(t *testing.T) {
	p := testGetProcess()

	n, err := p.NumThreads()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting NumThread error %v", err)
	assert.GreaterOrEqualf(t, n, int32(0), "invalid NumThread: %d", n)
}

func TestThreads(t *testing.T) {
	p := testGetProcess()

	n, err := p.NumThreads()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting NumThread error %v", err)
	assert.GreaterOrEqualf(t, n, int32(0), "invalid NumThread: %d", n)

	ts, err := p.Threads()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting Threads error %v", err)
	assert.Equalf(t, len(ts), int(n), "unexpected number of threads: %v vs %v", len(ts), n)
}

func TestName(t *testing.T) {
	p := testGetProcess()

	n, err := p.Name()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting name error %v", err)
	assert.Containsf(t, n, "process.test", "invalid Name %s", n)
}

// #nosec G204
func TestLong_Name_With_Spaces(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	require.NoErrorf(t, err, "unable to create temp dir %v", err)
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "loooong name with spaces.go")
	tmpfile, err := os.Create(tmpfilepath)
	require.NoErrorf(t, err, "unable to create temp file %v", err)

	tmpfilecontent := []byte("package main\nimport(\n\"time\"\n)\nfunc main(){\nfor range time.Tick(time.Second) {}\n}")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	require.NoErrorf(t, tmpfile.Close(), "unable to close temp file")
	ctx := context.Background()

	err = exec.CommandContext(ctx, "go", "build", "-o", tmpfile.Name()+".exe", tmpfile.Name()).Run() //nolint:gosec // test code
	require.NoErrorf(t, err, "unable to build temp file %v", err)

	cmd := exec.CommandContext(ctx, tmpfile.Name()+".exe") //nolint:gosec // test code

	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	n, err := p.Name()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting name error %v", err)
	basename := filepath.Base(tmpfile.Name() + ".exe")
	require.Equalf(t, basename, n, "%s != %s", basename, n)
	cmd.Process.Kill()
}

// #nosec G204
func TestLong_Name(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	require.NoErrorf(t, err, "unable to create temp dir %v", err)
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "looooooooooooooooooooong.go")
	tmpfile, err := os.Create(tmpfilepath)
	require.NoErrorf(t, err, "unable to create temp file %v", err)

	tmpfilecontent := []byte("package main\nimport(\n\"time\"\n)\nfunc main(){\nfor range time.Tick(time.Second) {}\n}")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	require.NoErrorf(t, tmpfile.Close(), "unable to close temp file")
	ctx := context.Background()

	err = exec.CommandContext(ctx, "go", "build", "-o", tmpfile.Name()+".exe", tmpfile.Name()).Run() //nolint:gosec // test code
	require.NoErrorf(t, err, "unable to build temp file %v", err)

	cmd := exec.CommandContext(ctx, tmpfile.Name()+".exe") //nolint:gosec // test code

	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	n, err := p.Name()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting name error %v", err)
	basename := filepath.Base(tmpfile.Name() + ".exe")
	require.Equalf(t, basename, n, "%s != %s", basename, n)
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
	ctx := context.Background()
	if out, err := exec.CommandContext(ctx, py3Path, "-c", "import psutil").CombinedOutput(); err != nil {
		t.Skipf("psutil not found for %s: %s", py3Path, out)
	}

	tmpdir, err := os.MkdirTemp("", "")
	require.NoErrorf(t, err, "unable to create temp dir %v", err)
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "looooooooooooooooooooong.py")
	tmpfile, err := os.Create(tmpfilepath)
	require.NoErrorf(t, err, "unable to create temp file %v", err)
	tmpfilecontent := []byte("#!" + py3Path + "\nimport psutil, time\nprint(psutil.Process().name(), flush=True)\nwhile True:\n\ttime.sleep(1)")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	require.NoErrorf(t, tmpfile.Chmod(0o744), "unable to chmod u+x temp file")
	require.NoErrorf(t, tmpfile.Close(), "unable to close temp file")
	cmd := exec.CommandContext(ctx, tmpfilepath)
	outPipe, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(outPipe)
	cmd.Start()
	defer cmd.Process.Kill()
	scanner.Scan()
	pyName := scanner.Text() // first line printed by py3 script, its name
	t.Logf("pyName %s", pyName)
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting process error %v", err)
	name, err := p.Name()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting name error %v", err)
	require.Equalf(t, pyName, name, "psutil and gopsutil process.Name() results differ: expected %s, got %s", pyName, name)
}

func TestExe(t *testing.T) {
	p := testGetProcess()

	n, err := p.Exe()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting Exe error %v", err)
	assert.Containsf(t, n, "process.test", "invalid Exe %s", n)
}

func TestCpuPercent(t *testing.T) {
	p := testGetProcess()
	_, err := p.Percent(0)
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	duration := time.Duration(1000) * time.Microsecond
	time.Sleep(duration)
	percent, err := p.Percent(0)
	require.NoError(t, err)

	numcpu := runtime.NumCPU()
	//	if percent < 0.0 || percent > 100.0*float64(numcpu) { // TODO
	require.GreaterOrEqualf(t, percent, 0.0, "CPUPercent value is invalid: %f, %d", percent, numcpu)
}

func TestCpuPercentLoop(t *testing.T) {
	p := testGetProcess()
	numcpu := runtime.NumCPU()

	for i := 0; i < 2; i++ {
		duration := time.Duration(100) * time.Microsecond
		percent, err := p.Percent(duration)
		if errors.Is(err, common.ErrNotImplementedError) {
			t.Skip("not implemented")
		}
		require.NoError(t, err)
		//	if percent < 0.0 || percent > 100.0*float64(numcpu) { // TODO
		require.GreaterOrEqualf(t, percent, 0.0, "CPUPercent value is invalid: %f, %d", percent, numcpu)
	}
}

func TestCreateTime(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skip CI")
	}

	p := testGetProcess()

	c, err := p.CreateTime()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	assert.GreaterOrEqualf(t, c, int64(1420000000), "process created time is wrong.")

	gotElapsed := time.Since(time.Unix(int64(c/1000), 0))
	maxElapsed := time.Duration(20 * time.Second)

	assert.Lessf(t, gotElapsed, maxElapsed, "this process has not been running for %v", gotElapsed)
}

func TestParent(t *testing.T) {
	p := testGetProcess()

	c, err := p.Parent()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	require.NotNilf(t, c, "could not get parent")
	require.NotZerof(t, c.Pid, "wrong parent pid")
}

func TestConnections(t *testing.T) {
	p := testGetProcess()
	ctx := context.Background()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0") // dynamically get a random open port from OS
	require.NoErrorf(t, err, "unable to resolve localhost: %v", err)
	l, err := net.ListenTCP(addr.Network(), addr)
	require.NoErrorf(t, err, "unable to listen on %v: %v", addr, err)
	defer l.Close()

	tcpServerAddr := l.Addr().String()
	tcpServerAddrIP := strings.Split(tcpServerAddr, ":")[0]
	tcpServerAddrPort, err := strconv.ParseUint(strings.Split(tcpServerAddr, ":")[1], 10, 32)
	require.NoErrorf(t, err, "unable to parse tcpServerAddr port: %v", err)

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
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", tcpServerAddr)
	require.NoErrorf(t, err, "unable to dial %v: %v", tcpServerAddr, err)
	defer conn.Close()

	// Rarely the call to net.Dial returns before the server connection is
	// established. Wait so that the test doesn't fail.
	<-serverEstablished

	c, err := p.Connections()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	require.NotEmptyf(t, c, "no connections found")

	serverConnections := 0
	for _, connection := range c {
		if connection.Laddr.IP == tcpServerAddrIP && uint64(connection.Laddr.Port) == tcpServerAddrPort && connection.Raddr.Port != 0 {
			require.Equalf(t, "ESTABLISHED", connection.Status, "expected server connection to be ESTABLISHED, have %+v", connection)
			serverConnections++
		}
	}

	clientConnections := 0
	for _, connection := range c {
		if connection.Raddr.IP == tcpServerAddrIP && uint64(connection.Raddr.Port) == tcpServerAddrPort {
			require.Equalf(t, "ESTABLISHED", connection.Status, "expected client connection to be ESTABLISHED, have %+v", connection)
			clientConnections++
		}
	}
	// two established connections, one for the server, the other for the client
	require.Equalf(t, 1, serverConnections, "expected 1 server connection, have %d.\nDetails: %+v", serverConnections, c)
	// two established connections, one for the server, the other for the client
	require.Equalf(t, 1, clientConnections, "expected 1 server connection, have %d.\nDetails: %+v", clientConnections, c)
}

func TestChildren(t *testing.T) {
	p := testGetProcess()
	ctx := context.Background()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "localhost", "-n", "4")
	} else {
		cmd = exec.CommandContext(ctx, "sleep", "3")
	}
	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)

	c, err := p.Children()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	require.NotEmptyf(t, c, "children is empty")
	found := false
	for _, child := range c {
		if child.Pid == int32(cmd.Process.Pid) {
			found = true
			break
		}
	}
	assert.Truef(t, found, "could not find child %d", cmd.Process.Pid)
}

func TestUsername(t *testing.T) {
	myPid := os.Getpid()
	currentUser, _ := user.Current()
	myUsername := currentUser.Username

	process, _ := NewProcess(int32(myPid))
	pidUsername, err := process.Username()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	assert.Equal(t, myUsername, pidUsername)

	t.Log(pidUsername)
}

func TestCPUTimes(t *testing.T) {
	pid := os.Getpid()
	process, err := NewProcess(int32(pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	spinSeconds := 0.2
	cpuTimes0, err := process.Times()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
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
	assert.Greaterf(t, measuredElapsed, float64(spinSeconds)/5, message)
	assert.Lessf(t, measuredElapsed, float64(spinSeconds)*5, message)
}

func TestOpenFiles(t *testing.T) {
	fp, err := os.Open("process_test.go")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, fp.Close())
	}()

	pid := os.Getpid()
	p, err := NewProcess(int32(pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	v, err := p.OpenFiles()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	assert.NotEmpty(t, v) // test always open files.

	for _, vv := range v {
		assert.NotEmpty(t, vv.Path)
	}
}

func TestKill(t *testing.T) {
	ctx := context.Background()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "localhost", "-n", "4")
	} else {
		cmd = exec.CommandContext(ctx, "sleep", "3")
	}
	require.NoError(t, cmd.Start())
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	err = p.Kill()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	cmd.Wait()
}

func TestIsRunning(t *testing.T) {
	ctx := context.Background()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "localhost", "-n", "2")
	} else {
		cmd = exec.CommandContext(ctx, "sleep", "1")
	}
	cmd.Start()
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)
	running, err := p.IsRunning()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "IsRunning error: %v", err)
	require.Truef(t, running, "process should be found running")
	cmd.Wait()
	running, err = p.IsRunning()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "IsRunning error: %v", err)
	require.Falsef(t, running, "process should NOT be found running")
}

// #nosec G204
func TestEnviron(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	require.NoErrorf(t, err, "unable to create temp dir %v", err)
	defer os.RemoveAll(tmpdir) // clean up
	tmpfilepath := filepath.Join(tmpdir, "test.go")
	tmpfile, err := os.Create(tmpfilepath)
	require.NoErrorf(t, err, "unable to create temp file %v", err)

	tmpfilecontent := []byte("package main\nimport(\n\"time\"\n)\nfunc main(){\nfor range time.Tick(time.Second) {}\n}")
	if _, err := tmpfile.Write(tmpfilecontent); err != nil {
		tmpfile.Close()
		t.Fatalf("unable to write temp file %v", err)
	}
	require.NoErrorf(t, tmpfile.Close(), "unable to close temp file")
	ctx := context.Background()

	err = exec.CommandContext(ctx, "go", "build", "-o", tmpfile.Name()+".exe", tmpfile.Name()).Run() //nolint:gosec // test code
	require.NoErrorf(t, err, "unable to build temp file %v", err)

	cmd := exec.CommandContext(ctx, tmpfile.Name()+".exe") //nolint:gosec // test code

	cmd.Env = []string{"testkey=envvalue"}

	require.NoError(t, cmd.Start())
	defer cmd.Process.Kill()
	time.Sleep(100 * time.Millisecond)
	p, err := NewProcess(int32(cmd.Process.Pid))
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoError(t, err)

	envs, err := p.Environ()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting environ error %v", err)
	var envvarFound bool
	if slices.Contains(envs, "testkey=envvalue") {
		envvarFound = true
	}
	assert.Truef(t, envvarFound, "environment variable not found")
}

func TestCwd(t *testing.T) {
	myPid := os.Getpid()
	currentWorkingDirectory, _ := os.Getwd()

	process, _ := NewProcess(int32(myPid))
	pidCwd, err := process.Cwd()
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
	require.NoErrorf(t, err, "getting cwd error %v", err)
	pidCwd = strings.TrimSuffix(pidCwd, string(os.PathSeparator))
	assert.Equal(t, currentWorkingDirectory, pidCwd)

	t.Log(pidCwd)
}

func TestConcurrent(t *testing.T) {
	const goroutines int = 5
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := NewProcess(int32(os.Getpid()))
			if err != nil {
				t.Errorf("NewProcess failed: %v", err)
				return
			}

			_, err = p.Times()
			if err != nil {
				t.Errorf("process.Times failed: %v", err)
				return
			}
		}()
	}
	wg.Wait()
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
