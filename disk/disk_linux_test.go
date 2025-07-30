// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package disk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFieldsOnMountinfo(t *testing.T) {
	fs := []string{"sysfs", "tmpfs"}

	lines := []string{
		"111 80 0:22 / /sys rw,nosuid,nodev,noexec,noatime shared:15 - sysfs sysfs rw",
		"114 80 0:61 / /run rw,nosuid,nodev shared:18 - tmpfs none rw,mode=755",
	}

	cases := map[string]struct {
		all    bool
		expect []PartitionStat
	}{
		"all": {
			all: true,
			expect: []PartitionStat{
				{Device: "sysfs", Mountpoint: "/sys", Fstype: "sysfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "noatime"}},
				{Device: "none", Mountpoint: "/run", Fstype: "tmpfs", Opts: []string{"rw", "nosuid", "nodev"}},
			},
		},
		"not all": {
			all: false,
			expect: []PartitionStat{
				{Device: "sysfs", Mountpoint: "/sys", Fstype: "sysfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "noatime"}},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := parseFieldsOnMountinfo(context.Background(), lines, c.all, fs, "")
			require.NoError(t, err)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func Test_parseFieldsOnMounts(t *testing.T) {
	fs := []string{"sysfs", "tmpfs"}

	lines := []string{
		"sysfs /sys sysfs rw,nosuid,nodev,noexec,noatime 0 0",
		"none /run tmpfs rw,nosuid,nodev,mode=755 0 0",
	}

	cases := map[string]struct {
		all    bool
		expect []PartitionStat
	}{
		"all": {
			all: true,
			expect: []PartitionStat{
				{Device: "sysfs", Mountpoint: "/sys", Fstype: "sysfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "noatime"}},
				{Device: "none", Mountpoint: "/run", Fstype: "tmpfs", Opts: []string{"rw", "nosuid", "nodev", "mode=755"}},
			},
		},
		"not all": {
			all: false,
			expect: []PartitionStat{
				{Device: "sysfs", Mountpoint: "/sys", Fstype: "sysfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "noatime"}},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := parseFieldsOnMounts(lines, c.all, fs)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestGetDeviceName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Controller notation conversion
		{"nvme0c0n1", "nvme0n1"},
		{"nvme10c23n1", "nvme10n1"},
		{"nvme5c5n2", "nvme5n2"},

		// Controller and partition together
		{"nvme0c0n1p1", "nvme0n1p1"},
		{"nvme2c2n1p2", "nvme2n1p2"},
		{"nvme10c23n1p3", "nvme10n1p3"},

		// Should NOT be changed
		{"nvme0n1", "nvme0n1"},     // standard notation
		{"nvme0n1p1", "nvme0n1p1"}, // partition
		{"sda", "sda"},             // non-nvme
		{"nvme5", "nvme5"},         // incomplete
		{"nvme0c0", "nvme0c0"},     // no namespace
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual := getDeviceName(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
