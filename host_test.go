package gopsutil

import (
	"testing"
)

func TestHostInfo(t *testing.T) {
	v, err := HostInfo()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Hostname == "" {
		t.Errorf("Could not get hostinfo %v", v)
	}
}

func TestBoot_time(t *testing.T) {
	v, err := BootTime()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v == 0 {
		t.Errorf("Could not boot time %v", v)
	}
}

func TestUsers(t *testing.T) {
	v, err := Users()
	if err != nil {
		t.Errorf("error %v", err)
	}
	for _, u := range v {
		if u.User == "" {
			t.Errorf("Could not Users %v", v)
		}
	}
}
