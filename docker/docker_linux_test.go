// +build linux

package docker

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
)

func TestGetDockerIDList(t *testing.T) {
	// If there is not docker environment, this test always fail.
	// not tested here
	/*
		_, err := GetDockerIDList()
		if err != nil {
			t.Errorf("error %v", err)
		}
	*/
}

func TestGetDockerStat(t *testing.T) {
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

func TestCgroupMountPoint(t *testing.T) {
	// see if docker executable is available
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("Docker is not available on this test machine")
	}
	b, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		t.Skip("Current test machine does not have /proc/mounts directory")
	}
	content := string(b)
	found := false
	targets := []string{"blkio", "cpu", "cpuacct", "cpuset", "devices", "freezer", "hugetlb", "memory", "net_cls", "net_prio", "perf_event", "pids", "systemd"}
	// if we find any targets that's good
	for _, target := range targets {
		mountPoint, err := cgroupMountPoint(target)
		if err != nil {
			continue
		}
		// test if cgroupMountPoint gets the mount point correctly
		if strings.Contains(content, fmt.Sprintf("cgroup %s cgroup", mountPoint)) {
			found = true
		}
	}
	if found == false {
		t.Errorf("Could not get any cgroup information!")
	}
}

func TestCgroupCPU(t *testing.T) {
	v, _ := GetDockerIDList()
	for _, id := range v {
		v, err := CgroupCPUDocker(id)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if v.CPU == "" {
			t.Errorf("could not get CgroupCPU %v", v)
		}

	}
}

func TestCgroupCPUInvalidId(t *testing.T) {
	_, err := CgroupCPUDocker("bad id")
	if err == nil {
		t.Error("Expected path does not exist error")
	}
}

func TestCgroupMem(t *testing.T) {
	v, _ := GetDockerIDList()
	for _, id := range v {
		v, err := CgroupMemDocker(id)
		if err != nil {
			t.Errorf("error %v", err)
		}
		empty := &CgroupMemStat{}
		if v == empty {
			t.Errorf("Could not CgroupMemStat %v", v)
		}
	}
}

func TestCgroupMemInvalidId(t *testing.T) {
	_, err := CgroupMemDocker("bad id")
	if err == nil {
		t.Error("Expected path does not exist error")
	}
}
