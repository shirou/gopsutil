// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package process_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/shirou/gopsutil/v4/process"
)

func TestTaskInfoMissingProcess(t *testing.T) {
	p := &process.Process{Pid: 999999}
	if got, err := p.TimesWithContext(context.Background()); err == nil {
		t.Errorf("Times returned success for missing process: %+v", got)
	}
	if got, err := p.NumThreadsWithContext(context.Background()); err == nil {
		t.Errorf("NumThreads returned success for missing process: %d", got)
	}
	if got, err := p.MemoryInfoWithContext(context.Background()); err == nil {
		t.Errorf("MemoryInfo returned success for missing process: %+v", got)
	}
}

func TestTaskInfoExitedProcess(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "/usr/bin/true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	p := &process.Process{Pid: int32(cmd.Process.Pid)}
	if got, err := p.Times(); !errors.Is(err, process.ErrorProcessNotRunning) || got != nil {
		t.Errorf("Times = %+v, %v", got, err)
	}
	if got, err := p.NumThreads(); !errors.Is(err, process.ErrorProcessNotRunning) || got != 0 {
		t.Errorf("NumThreads = %d, %v", got, err)
	}
	if got, err := p.MemoryInfo(); !errors.Is(err, process.ErrorProcessNotRunning) || got != nil {
		t.Errorf("MemoryInfo = %+v, %v", got, err)
	}
	if _, err := p.Percent(0); !errors.Is(err, process.ErrorProcessNotRunning) {
		t.Errorf("Percent error = %v", err)
	}
}

func TestTaskInfoCurrentProcess(t *testing.T) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := p.Times(); err != nil || got == nil {
		t.Fatalf("Times = %+v, %v", got, err)
	}
	if got, err := p.NumThreads(); err != nil || got <= 0 {
		t.Fatalf("NumThreads = %d, %v", got, err)
	}
	if got, err := p.MemoryInfo(); err != nil || got == nil || got.RSS == 0 {
		t.Fatalf("MemoryInfo = %+v, %v", got, err)
	}
}
