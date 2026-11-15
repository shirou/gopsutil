// SPDX-License-Identifier: BSD-3-Clause
//go:build linux || freebsd

package process

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func Test_SendSignal(t *testing.T) {
	checkPid := os.Getpid()

	p, _ := NewProcess(int32(checkPid))
	assert.NoErrorf(t, p.SendSignal(unix.SIGCONT), "send signal")
}

func TestGetTerminalMapPathsExist(t *testing.T) {
	termmap, err := getTerminalMap()
	if err != nil {
		t.Skipf("getTerminalMap not available: %v", err)
	}

	for _, name := range termmap {
		fullPath := common.HostDev(name)
		_, err := os.Stat(fullPath)
		assert.NoErrorf(t, err, "terminal device should exist: %s", fullPath)
	}
}

func TestGetTerminalMapKeepsTTYWhenPtsMissing(t *testing.T) {
	dev := t.TempDir()
	// Console tty entries without a pts directory — the FreeBSD-no-pty shape.
	for _, name := range []string{"ttyu0", "ttyv0"} {
		f, err := os.Create(filepath.Join(dev, name))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	t.Setenv("HOST_DEV", dev)

	termmap, err := getTerminalMap()
	if err != nil {
		t.Fatalf("getTerminalMap: %v", err)
	}
	if len(termmap) == 0 {
		t.Fatal("expected tty entries to remain when /dev/pts is missing")
	}
}

// Name only takes argv[0] when it extends the truncated comm, so a process
// with a long executable name cannot pick its own name through argv[0].
func TestNameCustomArgv0(t *testing.T) {
	if os.Getenv("GOPSUTIL_TEST_NAME_ARGV0") == "1" {
		fmt.Println("ready")
		_, _ = os.Stdin.Read(make([]byte, 1))
		os.Exit(0)
	}
	executable, err := os.Executable()
	require.NoError(t, err)
	data, err := os.ReadFile(executable)
	require.NoError(t, err)
	// Longer than comm on every platform, so the extended name path is taken.
	path := filepath.Join(t.TempDir(), "looooooooooooooooooooooooong")
	// #nosec G306 G703 -- fixed name under t.TempDir; the copied test binary must be executable.
	require.NoError(t, os.WriteFile(path, data, 0o700))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// #nosec G204 -- path is the test binary copied above.
	cmd := exec.CommandContext(ctx, path, "-test.run=^TestNameCustomArgv0$")
	cmd.Args[0] = "spoofed"
	cmd.Env = append(os.Environ(), "GOPSUTIL_TEST_NAME_ARGV0=1")
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())
	defer func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	}()
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "ready\n", ready)

	p, err := NewProcessWithContext(ctx, int32(cmd.Process.Pid))
	require.NoError(t, err)
	name, err := p.NameWithContext(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, "spoofed", name)
	assert.Truef(t, strings.HasPrefix(filepath.Base(path), name), "Name %q is not a prefix of %q", name, filepath.Base(path))
}
