// +build linux

package gopsutil

import (
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
