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
	lines := []string{
		"05   2 9:126 /           /              rw,noatime                      shared:1   - ext4  /dev/sda1 rw",
		"06   3 9:127 /           /foo           rw,noatime                      shared:1   - ext4  /dev/sda2 rw",
		"07   3 9:127 /bar        /foo/bar       rw,noatime                      shared:1   - ext4  /dev/sda2 rw", // "bind mount" on /foo
		"22  13 0:19  /           /dev/shm       rw,nosuid,nodev,noexec,relatime            - tmpfs -         rw",
		"37  29  0:4  net:[12345] /run/netns/foo rw                              shared:552 - nsfs  nsfs      rw",
		"111 80 0:22  /           /sys           rw,nosuid,nodev,noexec,noatime  shared:15  - sysfs sysfs     rw",
		"114 80 0:61  /           /run           rw,nosuid,nodev                 shared:18  - tmpfs none      rw,mode=755",
	}

	cases := map[string]struct {
		all    bool
		expect []PartitionStat
	}{
		"all": {
			all: true,
			expect: []PartitionStat{
				{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4", Opts: []string{"rw", "noatime"}},
				{Device: "/dev/sda2", Mountpoint: "/foo", Fstype: "ext4", Opts: []string{"rw", "noatime"}},
				{Device: "/foo", Mountpoint: "/foo/bar", Fstype: "ext4", Opts: []string{"rw", "noatime", "bind"}},
				{Device: "none", Mountpoint: "/dev/shm", Fstype: "tmpfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "relatime"}},
				{Device: "net:[12345]", Mountpoint: "/run/netns/foo", Fstype: "nsfs", Opts: []string{"rw"}},
				{Device: "none", Mountpoint: "/sys", Fstype: "sysfs", Opts: []string{"rw", "nosuid", "nodev", "noexec", "noatime"}},
				{Device: "none", Mountpoint: "/run", Fstype: "tmpfs", Opts: []string{"rw", "nosuid", "nodev"}},
			},
		},
		"not all": {
			all: false,
			expect: []PartitionStat{
				{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4", Opts: []string{"rw", "noatime"}},
				{Device: "/dev/sda2", Mountpoint: "/foo", Fstype: "ext4", Opts: []string{"rw", "noatime"}},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := parseFieldsOnMountinfo(context.Background(), lines, c.all, "")
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
