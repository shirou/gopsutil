// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package disk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func TestIOCountersSysfsSerialByDevice(t *testing.T) {
	root := t.TempDir()
	procDir := filepath.Join(root, "proc")
	sysDir := filepath.Join(root, "sys")
	runDir := filepath.Join(root, "run")
	for _, dir := range []string{procDir, filepath.Join(sysDir, "dev", "block"), runDir} {
		require.NoError(t, os.MkdirAll(dir, 0o755))
	}
	ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{
		common.HostProcEnvKey: procDir,
		common.HostSysEnvKey:  sysDir,
		common.HostRunEnvKey:  runDir,
		common.HostDevEnvKey:  filepath.Join(root, "no-device-nodes"),
	})

	// Symlinks mirror /sys/dev/block; partitions live under their owning disk.
	disks := []struct {
		name         string
		major, minor uint32
		partition    string
		partMinor    uint32
	}{
		{"sda", 8, 0, "sda1", 1},
		{"sdb", 8, 16, "sdb1", 17},
		{"nvme0n1", 259, 0, "nvme0n1p1", 1},
		{"nvme1n1", 259, 3, "nvme1n1p1", 4},
		{"mmcblk1", 179, 8, "mmcblk1p1", 9},
	}
	var diskstats []string
	want := make(map[string]string)
	linkDevice := func(major, minor uint32, target string) {
		t.Helper()
		link := filepath.Join(sysDir, "dev", "block", fmt.Sprintf("%d:%d", major, minor))
		rel, err := filepath.Rel(filepath.Dir(link), target)
		require.NoError(t, err)
		require.NoError(t, os.Symlink(rel, link))
	}
	for _, disk := range disks {
		dir := filepath.Join(sysDir, "devices", "test-controller", "block", disk.name)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "device"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "device", "model"), []byte(disk.name+"-model"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "device", "serial"), []byte(disk.name+"-serial"), 0o600))
		linkDevice(disk.major, disk.minor, dir)
		partDir := filepath.Join(dir, disk.partition)
		require.NoError(t, os.MkdirAll(partDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(partDir, "partition"), []byte("1\n"), 0o600))
		linkDevice(disk.major, disk.partMinor, partDir)
		for name, minor := range map[string]uint32{disk.name: disk.minor, disk.partition: disk.partMinor} {
			diskstats = append(diskstats, fmt.Sprintf("%d %d %s 1 0 2 0 3 0 4 0 0 5 6\n", disk.major, minor, name))
			want[name] = disk.name + "-model_" + disk.name + "-serial"
		}
	}
	require.NoError(t, os.WriteFile(filepath.Join(procDir, "diskstats"), []byte(strings.Join(diskstats, "")), 0o600))
	counters, err := IOCountersWithContext(ctx)
	require.NoError(t, err)
	require.Len(t, counters, len(want))
	for name, serial := range want {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, serial, counters[name].SerialNumber)
			require.Equal(t, uint64(1024), counters[name].ReadBytes)
			require.Equal(t, uint64(2048), counters[name].WriteBytes)
		})
	}
}

func TestSysfsSerialFallbackBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, model, serial, udev, want string
		linkPresent                     bool
	}{
		{"udev wins", "model", "serial", "E:ID_SERIAL=UDEV\n", "UDEV", true},
		{"udev without sysfs", "", "", "E:ID_SERIAL=UDEV\n", "UDEV", false},
		{"no model", "", "serial", "", "", true},
		{"no serial", "model", "", "", "", true},
		{"no device link", "model", "serial", "", "", false},
		{"untrimmed compatibility", "model\n", "serial\n", "", "model\n_serial\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			sysDir := filepath.Join(root, "sys")
			runDir := filepath.Join(root, "run")
			devDir := filepath.Join(sysDir, "dev", "block", "8:16", "device")
			require.NoError(t, os.MkdirAll(filepath.Join(runDir, "udev", "data"), 0o755))
			// An unrelated disk must never supply the requested disk's identity.
			wrongDir := filepath.Join(sysDir, "dev", "block", "8:0", "device")
			require.NoError(t, os.MkdirAll(wrongDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(wrongDir, "model"), []byte("WRONG"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(wrongDir, "serial"), []byte("WRONG"), 0o600))
			if tc.linkPresent {
				require.NoError(t, os.MkdirAll(devDir, 0o755))
				if tc.model != "" {
					require.NoError(t, os.WriteFile(filepath.Join(devDir, "model"), []byte(tc.model), 0o600))
				}
				if tc.serial != "" {
					require.NoError(t, os.WriteFile(filepath.Join(devDir, "serial"), []byte(tc.serial), 0o600))
				}
			}
			if tc.udev != "" {
				require.NoError(t, os.WriteFile(filepath.Join(runDir, "udev", "data", "b8:16"), []byte(tc.udev), 0o600))
			}
			ctx := context.WithValue(context.Background(), common.EnvKey, common.EnvMap{
				common.HostSysEnvKey: sysDir,
				common.HostRunEnvKey: runDir,
			})
			got, err := serialNumberFromDevNum(ctx, 8, 16)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
