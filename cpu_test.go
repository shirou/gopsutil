package gopsutil

import (
	"testing"
)

func TestCpu_times(t *testing.T) {
	v, err := Cpu_times(false)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(v) == 0 {
		t.Errorf("could not get CPUs ", err)
	}

	for _, vv := range v {
		if vv.User == 0 {
			t.Errorf("could not get CPU User: %v", vv)
		}
	}
}

func TestCpu_counts(t *testing.T) {
	v, err := Cpu_counts(true)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v == 0 {
		t.Errorf("could not get CPU counts: %v", v)
	}
}
