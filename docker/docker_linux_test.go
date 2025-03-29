// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
