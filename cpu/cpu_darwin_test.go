// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package cpu

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo_AppleSilicon(t *testing.T) {
	if runtime.GOARCH != "arm64" {
		t.Skip("wrong cpu type")
	}

	v, err := Info()
	require.NoErrorf(t, err, "cpu info should be implemented on darwin systems")

	for _, vv := range v {
		assert.NotEmptyf(t, vv.ModelName, "could not get CPU info: %v", vv)
		if vv.Mhz <= 0 && os.Getenv("CI") != "true" {
			t.Errorf("could not get frequency of: %s", vv.ModelName)
		}
		assert.LessOrEqualf(t, vv.Mhz, float64(6000), "cpu frequency is absurdly high value: %f MHz", vv.Mhz)
	}
}
