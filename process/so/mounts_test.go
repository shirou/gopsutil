// +build linux

package so

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMountInfo(t *testing.T) {
	buffer := bufio.NewReader(strings.NewReader(mountInfoTest))
	mountInfo := parseMountInfo(buffer)
	expected := []*mount{
		{dev: "0:53", root: "/", mountPoint: "/"},
		{dev: "0:55", root: "/", mountPoint: "/proc"},
		{dev: "0:56", root: "/", mountPoint: "/dev"},
		{dev: "0:57", root: "/", mountPoint: "/dev/pts"},
		{dev: "0:58", root: "/", mountPoint: "/sys"},
		{dev: "0:59", root: "/", mountPoint: "/sys/fs/cgroup"},
		{dev: "0:54", root: "/", mountPoint: "/dev/mqueue"},
		{dev: "0:60", root: "/", mountPoint: "/dev/shm"},
		{dev: "253:0", root: "/var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/resolv.conf", mountPoint: "/etc/resolv.conf"},
		{dev: "253:0", root: "/var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/hostname", mountPoint: "/etc/hostname"},
		{dev: "253:0", root: "/var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/hosts", mountPoint: "/etc/hosts"},
		{dev: "253:0", root: "/var/lib/docker/volumes/df3b1edbdd472155a90ce52613fd3231af9a3a161d370cd38876080fc89a5c72/_data", mountPoint: "/var/log/datadog"},
		{dev: "253:0", root: "/var/lib/docker/volumes/57abbcc4f7093acaeed6a255c761246ed01a2a63fceecd345a5258428b609ad5/_data", mountPoint: "/var/run/s6"},
	}

	assert.ElementsMatch(t, expected, mountInfo.mounts)
}

func TestGetMount(t *testing.T) {
	buffer := bufio.NewReader(strings.NewReader(mountInfoTest))
	mountInfo := parseMountInfo(buffer)

	type testCase struct {
		path          string
		expectedMount *mount
	}

	testCases := []testCase{
		{
			path: "/foobar",
			expectedMount: &mount{
				dev:        "0:53",
				root:       "/",
				mountPoint: "/",
			},
		},
		{
			path: "/etc/something",
			expectedMount: &mount{
				dev:        "0:53",
				root:       "/",
				mountPoint: "/",
			},
		},
		{
			path: "/etc/resolv.conf",
			expectedMount: &mount{
				dev:        "253:0",
				root:       "/var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/resolv.conf",
				mountPoint: "/etc/resolv.conf",
			},
		},
		{
			path: "/proc/1/maps",
			expectedMount: &mount{
				dev:        "0:55",
				root:       "/",
				mountPoint: "/proc",
			},
		},
		{
			path:          "bad-path",
			expectedMount: nil,
		},
	}

	for _, tc := range testCases {
		mount := mountInfo.GetMount(tc.path)
		assert.Equal(t, tc.expectedMount, mount)
	}

}

var mountInfoTest = `
599 553 0:53 / / rw,relatime master:211 - overlay overlay rw,lowerdir=/var/lib/docker/overlay2/l/773SKSZ22PXIPZWIAGH2GEIPVA:/var/lib/docker/overlay2/l/JHMVDLPEN3CKCT3ESIMMUHNUBK,upperdir=/var/lib/docker/overlay2/4d7c4a0bdb53df28926df1fce13a797e917c9fe60e8b5025cc961c36fb440203/diff,workdir=/var/lib/docker/overlay2/4d7c4a0bdb53df28926df1fce13a797e917c9fe60e8b5025cc961c36fb440203/work,xino=off
600 599 0:55 / /proc rw,nosuid,nodev,noexec,relatime - proc proc rw
601 599 0:56 / /dev rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755
602 601 0:57 / /dev/pts rw,nosuid,noexec,relatime - devpts devpts rw,gid=5,mode=620,ptmxmode=666
603 599 0:58 / /sys ro,nosuid,nodev,noexec,relatime - sysfs sysfs ro
604 603 0:59 / /sys/fs/cgroup rw,nosuid,nodev,noexec,relatime - tmpfs tmpfs rw,mode=755
622 601 0:54 / /dev/mqueue rw,nosuid,nodev,noexec,relatime - mqueue mqueue rw
623 601 0:60 / /dev/shm rw,nosuid,nodev,noexec,relatime - tmpfs shm rw,size=65536k
624 599 253:0 /var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/resolv.conf /etc/resolv.conf rw,relatime - ext4 /dev/mapper/vgvagrant-root rw,errors=remount-ro
625 599 253:0 /var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/hostname /etc/hostname rw,relatime - ext4 /dev/mapper/vgvagrant-root rw,errors=remount-ro
626 599 253:0 /var/lib/docker/containers/c97f9fa5c47e446eb49a7c78b2bd903fa41504e13b66a106eae1f80e141e1b9c/hosts /etc/hosts rw,relatime - ext4 /dev/mapper/vgvagrant-root rw,errors=remount-ro
627 599 253:0 /var/lib/docker/volumes/df3b1edbdd472155a90ce52613fd3231af9a3a161d370cd38876080fc89a5c72/_data /var/log/datadog rw,relatime master:1 - ext4 /dev/mapper/vgvagrant-root rw,errors=remount-ro
628 599 253:0 /var/lib/docker/volumes/57abbcc4f7093acaeed6a255c761246ed01a2a63fceecd345a5258428b609ad5/_data /var/run/s6 rw,relatime master:1 - ext4 /dev/mapper/vgvagrant-root rw,errors=remount-ro
`
