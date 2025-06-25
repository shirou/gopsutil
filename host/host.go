// SPDX-License-Identifier: BSD-3-Clause
package host

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type Warnings = common.Warnings

var invoke common.Invoker = common.Invoke{}

// A HostInfoStat describes the host status.
// This is not in the psutil but it useful.
type InfoStat struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	BootTime             uint64 `json:"bootTime"`
	Procs                uint64 `json:"procs"`           // number of processes
	OS                   string `json:"os"`              // ex: freebsd, linux
	Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
	PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion      string `json:"platformVersion"` // version of the complete OS
	KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
	KernelArch           string `json:"kernelArch"`      // native cpu architecture queried at runtime, as returned by `uname -m` or empty string in case of error
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"` // guest or host
	HostID               string `json:"hostId"`             // ex: uuid
}

type UserStat struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  int    `json:"started"`
}

func (h InfoStat) String() string {
	s, _ := json.Marshal(h)
	return string(s)
}

func (u UserStat) String() string {
	s, _ := json.Marshal(u)
	return string(s)
}

var enableBootTimeCache bool

// EnableBootTimeCache change cache behavior of BootTime. If true, cache BootTime value. Default is false.
func EnableBootTimeCache(enable bool) {
	enableBootTimeCache = enable
}

func Info() (*InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) (*InfoStat, error) {
	var err error
	ret := &InfoStat{
		OS: runtime.GOOS,
	}

	ret.Hostname, err = os.Hostname()
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting hostname: %w", err)
	}

	ret.Platform, ret.PlatformFamily, ret.PlatformVersion, err = PlatformInformationWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting platform information: %w", err)
	}

	ret.KernelVersion, err = KernelVersionWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting kernel version: %w", err)
	}

	ret.KernelArch, err = KernelArch()
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting kernel architecture: %w", err)
	}

	ret.VirtualizationSystem, ret.VirtualizationRole, err = VirtualizationWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting virtualization information: %w", err)
	}

	ret.BootTime, err = BootTimeWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting boot time: %w", err)
	}

	ret.Uptime, err = UptimeWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting uptime: %w", err)
	}

	ret.Procs, err = numProcs(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting number of procs: %w", err)
	}

	ret.HostID, err = HostIDWithContext(ctx)
	if err != nil && !errors.Is(err, common.ErrNotImplementedError) {
		return nil, fmt.Errorf("getting host ID: %w", err)
	}

	return ret, nil
}

// BootTime returns the system boot time expressed in seconds since the epoch.
func BootTime() (uint64, error) {
	return BootTimeWithContext(context.Background())
}

func Uptime() (uint64, error) {
	return UptimeWithContext(context.Background())
}

func Users() ([]UserStat, error) {
	return UsersWithContext(context.Background())
}

func PlatformInformation() (string, string, string, error) {
	return PlatformInformationWithContext(context.Background())
}

// HostID returns the unique host ID provided by the OS.
func HostID() (string, error) {
	return HostIDWithContext(context.Background())
}

func Virtualization() (string, string, error) {
	return VirtualizationWithContext(context.Background())
}

func KernelVersion() (string, error) {
	return KernelVersionWithContext(context.Background())
}

func timeSince(ts uint64) uint64 {
	return uint64(time.Now().Unix()) - ts
}

func timeSinceMillis(ts uint64) uint64 {
	return uint64(time.Now().UnixMilli()) - ts
}
