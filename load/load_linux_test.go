// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package load

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shirou/gopsutil/v4/common"
)

func TestFileAvgWithContextShortInput(t *testing.T) {
	for _, content := range []string{"", "\n", "0.25", "0.25 0.50"} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "loadavg"), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: dir})
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("short loadavg panicked: %v", r)
				}
			}()
			if value, err := fileAvgWithContext(ctx); err == nil || value != nil {
				t.Errorf("value=%v,error=%v;want nil,error", value, err)
			}
			value, err := AvgWithContext(ctx)
			if err != nil || value == nil {
				t.Errorf("sysinfo fallback failed:value=%v,error=%v", value, err)
			}
		})
	}
}

func TestFileAvgWithContextValues(t *testing.T) {
	for _, content := range []string{"0.25 0.50 1.75", "0.25 0.50 1.75 2/100 1234\n", "\t0.25\t0.50  1.75\n"} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "loadavg"), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: dir})
			value, err := fileAvgWithContext(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if value.Load1 != 0.25 || value.Load5 != 0.50 || value.Load15 != 1.75 {
				t.Errorf("unexpected values:%+v", value)
			}
		})
	}
}

func TestFileAvgWithContextReadAndParseErrors(t *testing.T) {
	for _, content := range []string{"bad 0.50 1.75", "0.25 bad 1.75", "0.25 0.50 bad"} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "loadavg"), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: dir})
			if value, err := fileAvgWithContext(ctx); err == nil || value != nil {
				t.Errorf("value=%v,error=%v;want nil,error", value, err)
			}
			if value, err := AvgWithContext(ctx); err != nil || value == nil {
				t.Errorf("fallback:value=%v,error=%v", value, err)
			}
		})
	}
	t.Run("missing", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{common.HostProcEnvKey: t.TempDir()})
		if value, err := fileAvgWithContext(ctx); err == nil || value != nil {
			t.Errorf("value=%v,error=%v;want nil,error", value, err)
		}
		if value, err := AvgWithContext(ctx); err != nil || value == nil {
			t.Errorf("fallback:value=%v,error=%v", value, err)
		}
	})
}
