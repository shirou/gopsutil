// SPDX-License-Identifier: BSD-3-Clause
package mem

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestVirtualMemory(t *testing.T) {
	if runtime.GOOS == "solaris" || runtime.GOOS == "illumos" {
		t.Skip("Only .Total .Available are supported on Solaris/illumos")
	}

	v, err := VirtualMemory()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	empty := &VirtualMemoryStat{}
	assert.NotSamef(t, v, empty, "error %v", v)
	t.Log(v)

	assert.Positive(t, v.Total)
	assert.Positive(t, v.Available)
	assert.Positive(t, v.Used)

	total := v.Used + v.Free + v.Buffers + v.Cached
	totalStr := "used + free + buffers + cached"
	switch runtime.GOOS {
	case "windows":
		total = v.Used + v.Available
		totalStr = "used + available"
	case "darwin", "openbsd":
		total = v.Used + v.Free + v.Cached + v.Inactive
		totalStr = "used + free + cached + inactive"
	case "freebsd":
		total = v.Used + v.Free + v.Cached + v.Inactive + v.Laundry
		totalStr = "used + free + cached + inactive + laundry"
	}
	assert.Equalf(t, v.Total, total,
		"Total should be computable (%v): %v", totalStr, v)

	assert.True(t, runtime.GOOS == "windows" || v.Free > 0)
	assert.Truef(t, runtime.GOOS == "windows" || v.Available > v.Free,
		"Free should be a subset of Available: %v", v)

	inDelta := assert.InDelta
	if runtime.GOOS == "windows" {
		inDelta = assert.InEpsilon
	}
	inDelta(t, v.UsedPercent,
		100*float64(v.Used)/float64(v.Total), 0.1,
		"UsedPercent should be how many percent of Total is Used: %v", v)
}

func TestSwapMemory(t *testing.T) {
	v, err := SwapMemory()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	empty := &SwapMemoryStat{}
	assert.NotSamef(t, v, empty, "error %v", v)

	t.Log(v)
}

func TestVirtualMemoryStat_String(t *testing.T) {
	v := VirtualMemoryStat{
		Total:       10,
		Available:   20,
		Used:        30,
		UsedPercent: 30.1,
		Free:        40,
	}
	t.Log(v)
	e := `{"total":10,"available":20,"used":30,"usedPercent":30.1,"free":40,"active":0,"inactive":0,"wired":0,"laundry":0,"buffers":0,"cached":0,"writeBack":0,"dirty":0,"writeBackTmp":0,"shared":0,"slab":0,"sreclaimable":0,"sunreclaim":0,"pageTables":0,"swapCached":0,"commitLimit":0,"committedAS":0,"highTotal":0,"highFree":0,"lowTotal":0,"lowFree":0,"swapTotal":0,"swapFree":0,"mapped":0,"vmallocTotal":0,"vmallocUsed":0,"vmallocChunk":0,"hugePagesTotal":0,"hugePagesFree":0,"hugePagesRsvd":0,"hugePagesSurp":0,"hugePageSize":0,"anonHugePages":0}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "VirtualMemoryStat string is invalid: %v", v)
}

func TestSwapMemoryStat_String(t *testing.T) {
	v := SwapMemoryStat{
		Total:       10,
		Used:        30,
		Free:        40,
		UsedPercent: 30.1,
		Sin:         1,
		Sout:        2,
		PgIn:        3,
		PgOut:       4,
		PgFault:     5,
		PgMajFault:  6,
	}
	e := `{"total":10,"used":30,"free":40,"usedPercent":30.1,"sin":1,"sout":2,"pgIn":3,"pgOut":4,"pgFault":5,"pgMajFault":6}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "SwapMemoryStat string is invalid: %v", v)
}

func TestSwapDevices(t *testing.T) {
	v, err := SwapDevices()
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "error calling SwapDevices: %v", err)

	t.Logf("SwapDevices() -> %+v", v)

	require.NotEmptyf(t, v, "no swap devices found. [this is expected if the host has swap disabled]")

	for _, device := range v {
		require.NotEmptyf(t, device.Name, "deviceName not set in %+v", device)
		if device.FreeBytes == 0 {
			t.Logf("[WARNING] free-bytes is zero in %+v. This might be expected", device)
		}
		if device.UsedBytes == 0 {
			t.Logf("[WARNING] used-bytes is zero in %+v. This might be expected", device)
		}
	}
}
