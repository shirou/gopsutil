// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package cpu

import (
	"os"
	"runtime"
	"testing"
)

func TestInfo_AppleSilicon(t *testing.T) {
	if runtime.GOARCH != "arm64" {
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
		if vv.Mhz <= 0 && os.Getenv("CI") != "true" {
			t.Errorf("could not get frequency of: %s", vv.ModelName)
		}
		if vv.Mhz > 6000 {
			t.Errorf("cpu frequency is absurdly high value: %f MHz", vv.Mhz)
		}
	}
}
