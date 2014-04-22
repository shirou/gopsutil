// +build linux

package gopsutil

import (
	"testing"
)

func TestLoad(t *testing.T) {
	load := NewLoad()

	v, err := load.LoadAvg()
	if err != nil {
		t.Errorf("error %v", err)
	}

	if v.Load1 == 0 || v.Load5 == 0 || v.Load15 == 0 {
		t.Errorf("error load: %v", v)
	}
}
