// +build android

package common

import (
	"context"
	"os/exec"
	"sync/atomic"
	"bufio"

	"golang.org/x/sys/unix"
)

func NumProcs() (uint64, error) {
	ps, err := exec.LookPath("ps")
	if err != nil {
		return 0, err
	}
	cmd := exec.Command(ps, "-A", "-o", "PID")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	var cnt uint64

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		cnt++
	}

	if err := scanner.Err(); err != nil {
		return cnt, err
	}

	if err := cmd.Wait(); err != nil {
		return 0, err
	}

	return cnt, err
}

// cachedBootTime must be accessed via atomic.Load/StoreUint64
var cachedBootTime uint64

func BootTimeWithContext(ctx context.Context) (uint64, error) {
	t := atomic.LoadUint64(&cachedBootTime)
	if t != 0 {
		return t, nil
	}

	sysinfo := &unix.Sysinfo_t{}

	if err := unix.Sysinfo(sysinfo); err != nil {
		return 0, err
	}

	uptime := uint64(sysinfo.Uptime)
	atomic.StoreUint64(&uptime, t)
	return uptime, nil
}

func Virtualization() (string, string, error) {
	return VirtualizationWithContext(context.Background())
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	return "", "", nil
}

func GetOSRelease() (platform string, version string, err error) {
	getprop, err := exec.LookPath("getprop")
	if err != nil {
		return "", "", err
	}
	flavorCmd := exec.Command(getprop, "ro.build.flavor")
	platformProp, err := flavorCmd.Output()

	if err != nil {
		return "", "", err
	}

	versionCmd := exec.Command(getprop, "ro.build.version.release")
	versionProp, err := versionCmd.Output()

	if err != nil {
		return "", "", err
	}

	return string(platformProp), string(versionProp), nil
}
