// SPDX-License-Identifier: BSD-3-Clause
//go:build plan9

package mem

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

var virtualMemoryTests = []struct {
	mockedRootFS string
	stat         *VirtualMemoryStat
}{
	{
		"swap", &VirtualMemoryStat{
			Total:       1071185920,
			Available:   808370176,
			Used:        11436032,
			UsedPercent: 1.3949677238843257,
			Free:        808370176,
			SwapTotal:   655360000,
			SwapFree:    655360000,
		},
	},
}

func TestVirtualMemoryPlan9(t *testing.T) {
	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			t.Setenv("HOST_ROOT", "testdata/plan9/virtualmemory/")

			stat, err := VirtualMemory()
			common.SkipIfNotImplementedErr(t, err)
			require.NoError(t, err)
			assert.Truef(t, reflect.DeepEqual(stat, tt.stat), "got: %+v\nwant: %+v", stat, tt.stat)
		})
	}
}

var swapMemoryTests = []struct {
	mockedRootFS string
	swap         *SwapMemoryStat
}{
	{
		"swap", &SwapMemoryStat{
			Total: 655360000,
			Used:  0,
			Free:  655360000,
		},
	},
}

func TestSwapMemoryPlan9(t *testing.T) {
	for _, tt := range swapMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			t.Setenv("HOST_ROOT", "testdata/plan9/virtualmemory/")

			swap, err := SwapMemory()
			common.SkipIfNotImplementedErr(t, err)
			require.NoError(t, err)
			assert.Truef(t, reflect.DeepEqual(swap, tt.swap), "got: %+v\nwant: %+v", swap, tt.swap)
		})
	}
}
