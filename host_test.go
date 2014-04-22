package gopsutil

import (
	"testing"
)

func TestHostInfo(t *testing.T) {
	host := NewHost()

	v, err := host.HostInfo()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Uptime == 0 {
		t.Errorf("Could not get uptime %v", v)
	}
}
