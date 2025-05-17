// SPDX-License-Identifier: BSD-3-Clause
package cpu

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestTimes(t *testing.T) {
	v, err := Times(false)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, v, "could not get CPUs: %s", err)
	empty := TimesStat{}
	for _, vv := range v {
		assert.NotEqualf(t, vv, empty, "could not get CPU User: %v", vv)
	}

	// test sum of per cpu stats is within margin of error for cpu total stats
	cpuTotal, err := Times(false)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, cpuTotal, "could not get CPUs: %s", err)
	perCPU, err := Times(true)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, perCPU, "could not get CPUs: %s", err)
	var perCPUUserTimeSum float64
	var perCPUSystemTimeSum float64
	var perCPUIdleTimeSum float64
	for _, pc := range perCPU {
		perCPUUserTimeSum += pc.User
		perCPUSystemTimeSum += pc.System
		perCPUIdleTimeSum += pc.Idle
	}
	margin := 2.0
	t.Log(cpuTotal[0])

	if cpuTotal[0].User == 0 && cpuTotal[0].System == 0 && cpuTotal[0].Idle == 0 {
		t.Error("could not get cpu values")
	}
	if cpuTotal[0].User != 0 {
		assert.InEpsilon(t, cpuTotal[0].User, perCPUUserTimeSum, margin)
	}
	if cpuTotal[0].System != 0 {
		assert.InEpsilon(t, cpuTotal[0].System, perCPUSystemTimeSum, margin)
	}
	if cpuTotal[0].Idle != 0 {
		assert.InEpsilon(t, cpuTotal[0].Idle, perCPUIdleTimeSum, margin)
	}
}

func TestCounts(t *testing.T) {
	logicalCount, err := Counts(true)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotZerof(t, logicalCount, "could not get logical CPU counts: %v", logicalCount)
	t.Logf("logical cores: %d", logicalCount)
	physicalCount, err := Counts(false)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotZerof(t, physicalCount, "could not get physical CPU counts: %v", physicalCount)
	t.Logf("physical cores: %d", physicalCount)
	assert.GreaterOrEqualf(t, logicalCount, physicalCount, "logical CPU count should be greater than or equal to physical CPU count: %v >= %v", logicalCount, physicalCount)
}

func TestTimeStat_String(t *testing.T) {
	v := TimesStat{
		CPU:    "cpu0",
		User:   100.1,
		System: 200.1,
		Idle:   300.1,
	}
	e := `{"cpu":"cpu0","user":100.1,"system":200.1,"idle":300.1,"nice":0.0,"iowait":0.0,"irq":0.0,"softirq":0.0,"steal":0.0,"guest":0.0,"guestNice":0.0}`
	assert.JSONEqf(t, e, v.String(), "CPUTimesStat string is invalid: %v", v)
}

func TestInfo(t *testing.T) {
	v, err := Info()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, v, "could not get CPU Info")
	for _, vv := range v {
		assert.NotEmptyf(t, vv.ModelName, "could not get CPU Info: %v", vv)
	}
}

func testPercent(t *testing.T, percpu bool) {
	t.Helper()
	numcpu := runtime.NumCPU()
	testCount := 3

	if runtime.GOOS != "windows" {
		testCount = 100
		v, err := Percent(time.Millisecond, percpu)
		common.SkipIfNotImplementedErr(t, err)
		require.NoError(t, err)
		// Skip CI which CPU num is different
		if os.Getenv("CI") != "true" {
			if (percpu && len(v) != numcpu) || (!percpu && len(v) != 1) {
				t.Fatalf("wrong number of entries from CPUPercent: %v", v)
			}
		}
	}
	for i := 0; i < testCount; i++ {
		duration := time.Duration(10) * time.Microsecond
		v, err := Percent(duration, percpu)
		common.SkipIfNotImplementedErr(t, err)
		require.NoError(t, err)
		for _, percent := range v {
			// Check for slightly greater then 100% to account for any rounding issues.
			if percent < 0.0 || percent > 100.0001*float64(numcpu) {
				t.Fatalf("CPUPercent value is invalid: %f", percent)
			}
		}
	}
}

func testPercentLastUsed(t *testing.T, percpu bool) {
	t.Helper()
	numcpu := runtime.NumCPU()
	testCount := 10

	if runtime.GOOS != "windows" {
		testCount = 2
		v, err := Percent(time.Millisecond, percpu)
		common.SkipIfNotImplementedErr(t, err)
		require.NoError(t, err)
		// Skip CI which CPU num is different
		if os.Getenv("CI") != "true" {
			if (percpu && len(v) != numcpu) || (!percpu && len(v) != 1) {
				t.Fatalf("wrong number of entries from CPUPercent: %v", v)
			}
		}
	}
	for i := 0; i < testCount; i++ {
		v, err := Percent(0, percpu)
		common.SkipIfNotImplementedErr(t, err)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
		for _, percent := range v {
			// Check for slightly greater then 100% to account for any rounding issues.
			if percent < 0.0 || percent > 100.0001*float64(numcpu) {
				t.Fatalf("CPUPercent value is invalid: %f", percent)
			}
		}
	}
}

func TestPercent(t *testing.T) {
	testPercent(t, false)
}

func TestPercentPerCpu(t *testing.T) {
	testPercent(t, true)
}

func TestPercentIntervalZero(t *testing.T) {
	testPercentLastUsed(t, false)
}

func TestPercentIntervalZeroPerCPU(t *testing.T) {
	testPercentLastUsed(t, true)
}
