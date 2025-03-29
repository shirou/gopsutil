// SPDX-License-Identifier: BSD-3-Clause
//go:build plan9

package cpu

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var timesTests = []struct {
	mockedRootFS string
	stats        []TimesStat
}{
	{
		"2cores",
		[]TimesStat{
			{
				CPU:    "Core i7/Xeon",
				User:   2780.0 / 1000.0,
				System: 30020.0 / 1000.0,
				Idle:   (1412961713341830*2)/1000000000.0 - 2.78 - 30.02,
			},
		},
	},
}

func TestTimesPlan9(t *testing.T) {
	for _, tt := range timesTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			t.Setenv("HOST_ROOT", filepath.Join("testdata/plan9", tt.mockedRootFS))
			stats, err := Times(false)
			common.SkipIfNotImplementedErr(t, err)
			require.NoError(t, err)
			eps := cmpopts.EquateApprox(0, 0.00000001)
			assert.Truef(t, cmp.Equal(stats, tt.stats, eps), "got: %+v\nwant: %+v", stats, tt.stats)
		})
	}
}
