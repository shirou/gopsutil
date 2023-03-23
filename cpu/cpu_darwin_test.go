//go:build darwin
// +build darwin

package cpu

import (
	"testing"

	"github.com/shoenig/go-m1cpu"
)

func Test_CpuInfo_AppleSilicon(t *testing.T) {
	if !m1cpu.IsAppleSilicon() {
		t.Skip("wrong cpu type")
	}

	v, err := Info()
	if err != nil {
		t.Errorf("cpu info should be implemented on darwin systems")
	}

	for _, vv := range v {
		if vv.ModelName == "" {
			t.Errorf("could not get CPU info: %v", vv)
		}
		if vv.Mhz <= 0 {
			t.Errorf("could not get frequency of: %s", vv.ModelName)
		}
		if vv.Mhz > 6000 {
			t.Errorf("cpu frequency is absurdly high value: %f MHz", vv.Mhz)
		}
	}
}
