// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package docker

import (
	"context"
	"math"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	// testCgroupfsID lives under <cgroup root>/docker/<id> (cgroupfs driver)
	// in both testdata hierarchies.
	testCgroupfsID = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	// testSystemdID lives under <cgroup root>/system.slice/docker-<id>.scope
	// (systemd driver) in both testdata hierarchies.
	testSystemdID = "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
)

// setTestdataHostSys points HOST_SYS at the testdata tree of the given cgroup
// hierarchy ("cgroup1" or "cgroup2") and returns that path.
func setTestdataHostSys(t *testing.T, hierarchy string) string {
	t.Helper()
	hostSys := filepath.Join("testdata", "linux", hierarchy, "sys")
	t.Setenv("HOST_SYS", hostSys)
	return hostSys
}

func TestGetDockerIDList(_ *testing.T) {
	// If there is not docker environment, this test always fail.
	// not tested here
	/*
		_, err := GetDockerIDList()
		if err != nil {
			t.Errorf("error %v", err)
		}
	*/
}

func TestGetDockerStat(_ *testing.T) {
	// If there is not docker environment, this test always fail.
	// not tested here

	/*
		ret, err := GetDockerStat()
		if err != nil {
			t.Errorf("error %v", err)
		}
		if len(ret) == 0 {
			t.Errorf("ret is empty")
		}
		empty := CgroupDockerStat{}
		for _, v := range ret {
			if empty == v {
				t.Errorf("empty CgroupDockerStat")
			}
			if v.ContainerID == "" {
				t.Errorf("Could not get container id")
			}
		}
	*/
}

func TestCgroupCPU(t *testing.T) {
	v, _ := GetDockerIDList()
	for _, id := range v {
		v, err := CgroupCPUDockerWithContext(context.Background(), id)
		require.NoError(t, err)
		assert.NotEmptyf(t, v.CPU, "could not get CgroupCPU %v", v)

	}
}

func TestCgroupCPUInvalidId(t *testing.T) {
	_, err := CgroupCPUDockerWithContext(context.Background(), "bad id")
	assert.Errorf(t, err, "Expected path does not exist error")
}

func TestCgroupMem(t *testing.T) {
	v, _ := GetDockerIDList()
	for _, id := range v {
		v, err := CgroupMemDocker(id)
		require.NoError(t, err)
		empty := &CgroupMemStat{}
		assert.NotSamef(t, v, empty, "Could not CgroupMemStat %v", v)
	}
}

func TestCgroupMemInvalidId(t *testing.T) {
	_, err := CgroupMemDocker("bad id")
	assert.Errorf(t, err, "Expected path does not exist error")
}

func TestCgroupCPUTestdata(t *testing.T) {
	tests := []struct {
		name        string
		hierarchy   string
		containerID string
		user        float64
		system      float64
		usage       float64
	}{
		{ // https://github.com/shirou/gopsutil/issues/1416
			name:        "cgroup v2 cgroupfs driver",
			hierarchy:   "cgroup2",
			containerID: testCgroupfsID,
			user:        745.383689,
			system:      315.767382,
			usage:       1061.151072,
		},
		{ // https://github.com/shirou/gopsutil/issues/1416
			name:        "cgroup v2 systemd driver",
			hierarchy:   "cgroup2",
			containerID: testSystemdID,
			user:        1.5,
			system:      1.0,
			usage:       2.5,
		},
		{
			name:        "cgroup v1 cgroupfs driver",
			hierarchy:   "cgroup1",
			containerID: testCgroupfsID,
			user:        12345 / cpu.ClocksPerSec,
			system:      6789 / cpu.ClocksPerSec,
			usage:       1061.151072,
		},
		{
			name:        "cgroup v1 systemd driver",
			hierarchy:   "cgroup1",
			containerID: testSystemdID,
			user:        150000 / cpu.ClocksPerSec,
			system:      100000 / cpu.ClocksPerSec,
			usage:       2.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestdataHostSys(t, tt.hierarchy)

			stat, err := CgroupCPUDockerWithContext(t.Context(), tt.containerID)
			require.NoError(t, err)
			assert.Equal(t, tt.containerID, stat.CPU)
			assert.InDelta(t, tt.user, stat.User, 1e-9)
			assert.InDelta(t, tt.system, stat.System, 1e-9)
			assert.InDelta(t, tt.usage, stat.Usage, 1e-9)

			usage, err := CgroupCPUDockerUsageWithContext(t.Context(), tt.containerID)
			require.NoError(t, err)
			assert.InDelta(t, tt.usage, usage, 1e-9)
		})
	}
}

func TestCgroupMemTestdata(t *testing.T) {
	tests := []struct {
		name        string
		hierarchy   string
		containerID string
		want        *CgroupMemStat
	}{
		{ // https://github.com/shirou/gopsutil/issues/1416
			name:        "cgroup v2 cgroupfs driver",
			hierarchy:   "cgroup2",
			containerID: testCgroupfsID,
			want: &CgroupMemStat{
				ContainerID:        testCgroupfsID,
				Cache:              14278656,
				RSS:                10895360,
				RSSHuge:            2097152,
				MappedFile:         9904128,
				Pgfault:            21863145,
				Pgmajfault:         544,
				InactiveAnon:       446464,
				ActiveAnon:         10457088,
				InactiveFile:       2482176,
				ActiveFile:         11796480,
				Unevictable:        4096,
				TotalCache:         14278656,
				TotalRSS:           10895360,
				TotalRSSHuge:       2097152,
				TotalMappedFile:    9904128,
				TotalPgFault:       21863145,
				TotalPgMajFault:    544,
				TotalInactiveAnon:  446464,
				TotalActiveAnon:    10457088,
				TotalInactiveFile:  2482176,
				TotalActiveFile:    11796480,
				TotalUnevictable:   4096,
				MemUsageInBytes:    26591232,
				MemMaxUsageInBytes: 31875072,
				MemLimitInBytes:    math.MaxUint64, // memory.max is "max"
				MemFailCnt:         3,
			},
		},
		{ // https://github.com/shirou/gopsutil/issues/1416
			name:        "cgroup v2 systemd driver without memory.peak",
			hierarchy:   "cgroup2",
			containerID: testSystemdID,
			want: &CgroupMemStat{
				ContainerID:        testSystemdID,
				Cache:              8388608,
				RSS:                4194304,
				RSSHuge:            0,
				MappedFile:         1048576,
				Pgfault:            1000,
				Pgmajfault:         10,
				InactiveAnon:       1048576,
				ActiveAnon:         3145728,
				InactiveFile:       2097152,
				ActiveFile:         6291456,
				Unevictable:        0,
				TotalCache:         8388608,
				TotalRSS:           4194304,
				TotalRSSHuge:       0,
				TotalMappedFile:    1048576,
				TotalPgFault:       1000,
				TotalPgMajFault:    10,
				TotalInactiveAnon:  1048576,
				TotalActiveAnon:    3145728,
				TotalInactiveFile:  2097152,
				TotalActiveFile:    6291456,
				TotalUnevictable:   0,
				MemUsageInBytes:    123456789,
				MemMaxUsageInBytes: 0, // memory.peak needs Linux 5.19+
				MemLimitInBytes:    268435456,
				MemFailCnt:         7,
			},
		},
		{
			name:        "cgroup v1 cgroupfs driver",
			hierarchy:   "cgroup1",
			containerID: testCgroupfsID,
			want: &CgroupMemStat{
				ContainerID:             testCgroupfsID,
				Cache:                   14278656,
				RSS:                     10895360,
				RSSHuge:                 2097152,
				MappedFile:              9904128,
				Pgpgin:                  5000,
				Pgpgout:                 4000,
				Pgfault:                 21863145,
				Pgmajfault:              544,
				InactiveAnon:            446464,
				ActiveAnon:              10457088,
				InactiveFile:            2482176,
				ActiveFile:              11796480,
				Unevictable:             4096,
				HierarchicalMemoryLimit: 9223372036854771712,
				TotalCache:              24278656,
				TotalRSS:                20895360,
				TotalRSSHuge:            4097152,
				TotalMappedFile:         19904128,
				TotalPgpgIn:             15000,
				TotalPgpgOut:            14000,
				TotalPgFault:            31863145,
				TotalPgMajFault:         1544,
				TotalInactiveAnon:       1446464,
				TotalActiveAnon:         20457088,
				TotalInactiveFile:       12482176,
				TotalActiveFile:         21796480,
				TotalUnevictable:        14096,
				MemUsageInBytes:         26591232,
				MemMaxUsageInBytes:      31875072,
				MemLimitInBytes:         9223372036854771712,
				MemFailCnt:              0,
			},
		},
		{
			name:        "cgroup v1 systemd driver without memory.max_usage_in_bytes",
			hierarchy:   "cgroup1",
			containerID: testSystemdID,
			want: &CgroupMemStat{
				ContainerID:             testSystemdID,
				Cache:                   8388608,
				RSS:                     4194304,
				RSSHuge:                 0,
				MappedFile:              1048576,
				Pgpgin:                  300,
				Pgpgout:                 200,
				Pgfault:                 1000,
				Pgmajfault:              10,
				InactiveAnon:            1048576,
				ActiveAnon:              3145728,
				InactiveFile:            2097152,
				ActiveFile:              6291456,
				Unevictable:             0,
				HierarchicalMemoryLimit: 268435456,
				TotalCache:              8388608,
				TotalRSS:                4194304,
				TotalRSSHuge:            0,
				TotalMappedFile:         1048576,
				TotalPgpgIn:             300,
				TotalPgpgOut:            200,
				TotalPgFault:            1000,
				TotalPgMajFault:         10,
				TotalInactiveAnon:       1048576,
				TotalActiveAnon:         3145728,
				TotalInactiveFile:       2097152,
				TotalActiveFile:         6291456,
				TotalUnevictable:        0,
				MemUsageInBytes:         123456789,
				MemMaxUsageInBytes:      0,
				MemLimitInBytes:         268435456,
				MemFailCnt:              7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setTestdataHostSys(t, tt.hierarchy)

			got, err := CgroupMemDockerWithContext(t.Context(), tt.containerID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestCgroupExplicitBase checks that a caller-supplied base directory is
// honoured verbatim on both hierarchies, including the documented
// systemd.slice usage (containerID = docker-<id>.scope).
func TestCgroupExplicitBase(t *testing.T) {
	tests := []struct {
		name        string
		hierarchy   string
		containerID string
		cpuBase     []string
		memBase     []string
	}{
		{
			name:        "cgroup v2 cgroupfs driver",
			hierarchy:   "cgroup2",
			containerID: testCgroupfsID,
			cpuBase:     []string{"fs", "cgroup", "docker"},
			memBase:     []string{"fs", "cgroup", "docker"},
		},
		{
			name:        "cgroup v2 systemd scope",
			hierarchy:   "cgroup2",
			containerID: "docker-" + testSystemdID + ".scope",
			cpuBase:     []string{"fs", "cgroup", "system.slice"},
			memBase:     []string{"fs", "cgroup", "system.slice"},
		},
		{
			name:        "cgroup v1 cgroupfs driver",
			hierarchy:   "cgroup1",
			containerID: testCgroupfsID,
			cpuBase:     []string{"fs", "cgroup", "cpuacct", "docker"},
			memBase:     []string{"fs", "cgroup", "memory", "docker"},
		},
		{
			name:        "cgroup v1 systemd scope",
			hierarchy:   "cgroup1",
			containerID: "docker-" + testSystemdID + ".scope",
			cpuBase:     []string{"fs", "cgroup", "cpuacct", "system.slice"},
			memBase:     []string{"fs", "cgroup", "memory", "system.slice"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostSys := setTestdataHostSys(t, tt.hierarchy)
			cpuBase := filepath.Join(append([]string{hostSys}, tt.cpuBase...)...)
			memBase := filepath.Join(append([]string{hostSys}, tt.memBase...)...)

			stat, err := CgroupCPUWithContext(t.Context(), tt.containerID, cpuBase)
			require.NoError(t, err)
			assert.Equal(t, tt.containerID, stat.CPU)
			assert.Positive(t, stat.Usage)

			usage, err := CgroupCPUUsageWithContext(t.Context(), tt.containerID, cpuBase)
			require.NoError(t, err)
			assert.InDelta(t, stat.Usage, usage, 1e-9)

			mem, err := CgroupMemWithContext(t.Context(), tt.containerID, memBase)
			require.NoError(t, err)
			assert.Equal(t, tt.containerID, mem.ContainerID)
			assert.Positive(t, mem.MemUsageInBytes)
			assert.Positive(t, mem.MemLimitInBytes)
		})
	}
}

func TestCgroupInvalidContainerID(t *testing.T) {
	tests := []struct {
		name        string
		containerID string
	}{
		{name: "parent traversal", containerID: "../escape"},
		{name: "path separator", containerID: "a/b"},
		{name: "backslash", containerID: `a\b`},
		{name: "unknown id", containerID: "0000"},
	}
	for _, hierarchy := range []string{"cgroup1", "cgroup2"} {
		for _, tt := range tests {
			t.Run(hierarchy+" "+tt.name, func(t *testing.T) {
				setTestdataHostSys(t, hierarchy)

				_, err := CgroupCPUDockerWithContext(t.Context(), tt.containerID)
				require.Error(t, err)
				_, err = CgroupCPUDockerUsageWithContext(t.Context(), tt.containerID)
				require.Error(t, err)
				_, err = CgroupMemDockerWithContext(t.Context(), tt.containerID)
				require.Error(t, err)
			})
		}
	}
}
