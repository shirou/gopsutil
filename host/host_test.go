// SPDX-License-Identifier: BSD-3-Clause
package host

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestHostID(t *testing.T) {
	v, err := HostID()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, v, "Could not get host id %v", v)
	t.Log(v)
}

func TestInfo(t *testing.T) {
	v, err := Info()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	empty := &InfoStat{}
	assert.NotSamef(t, v, empty, "Could not get hostinfo %v", v)
	assert.NotZerof(t, v.Procs, "Could not determine the number of host processes")
	t.Log(v)
}

func TestUptime(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skip CI")
	}

	v, err := Uptime()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotZerof(t, v, "Could not get up time %v", v)
}

func TestBootTime(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skip CI")
	}
	v, err := BootTime()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotZerof(t, v, "Could not get boot time %v", v)
	assert.GreaterOrEqualf(t, v, uint64(946652400), "Invalid Boottime, older than 2000-01-01")
	t.Logf("first boot time: %d", v)

	v2, err := BootTime()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.Equalf(t, v, v2, "cached boot time is different")
	t.Logf("second boot time: %d", v2)
}

func TestUsers(t *testing.T) {
	v, err := Users()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	empty := UserStat{}
	if len(v) == 0 {
		t.Skip("Users is empty")
	}
	for _, u := range v {
		assert.NotEqualf(t, u, empty, "Could not Users %v", v)
		t.Log(u)
	}
}

func TestInfoStat_String(t *testing.T) {
	v := InfoStat{
		Hostname:   "test",
		Uptime:     3000,
		Procs:      100,
		OS:         "linux",
		Platform:   "ubuntu",
		BootTime:   1447040000,
		HostID:     "edfd25ff-3c9c-b1a4-e660-bd826495ad35",
		KernelArch: "x86_64",
	}
	e := `{"hostname":"test","uptime":3000,"bootTime":1447040000,"procs":100,"os":"linux","platform":"ubuntu","platformFamily":"","platformVersion":"","kernelVersion":"","kernelArch":"x86_64","virtualizationSystem":"","virtualizationRole":"","hostId":"edfd25ff-3c9c-b1a4-e660-bd826495ad35"}`
	assert.JSONEqf(t, e, v.String(), "HostInfoStat string is invalid:\ngot  %v\nwant %v", v, e)
}

func TestUserStat_String(t *testing.T) {
	v := UserStat{
		User:     "user",
		Terminal: "term",
		Host:     "host",
		Started:  100,
	}
	e := `{"user":"user","terminal":"term","host":"host","started":100}`
	assert.JSONEqf(t, e, v.String(), "UserStat string is invalid: %v", v)
}

func TestGuid(t *testing.T) {
	id, err := HostID()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, id, "Host id is empty")
	t.Logf("Host id value: %v", id)
}

func TestVirtualization(t *testing.T) {
	wg := sync.WaitGroup{}
	testCount := 10
	wg.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func(j int) {
			system, role, err := Virtualization()
			defer wg.Done()
			common.SkipIfNotImplementedErr(t, err)
			assert.NoErrorf(t, err, "Virtualization() failed, %v", err)

			if j == 9 {
				t.Logf("Virtualization(): %s, %s", system, role)
			}
		}(i)
	}
	wg.Wait()
}

func TestKernelVersion(t *testing.T) {
	version, err := KernelVersion()
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "KernelVersion() failed, %v", err)
	assert.NotEmptyf(t, version, "KernelVersion() returns empty: %s", version)

	t.Logf("KernelVersion(): %s", version)
}

func TestPlatformInformation(t *testing.T) {
	platform, family, version, err := PlatformInformation()
	common.SkipIfNotImplementedErr(t, err)
	require.NoErrorf(t, err, "PlatformInformation() failed, %v", err)
	assert.NotEmptyf(t, platform, "PlatformInformation() returns empty: %v", platform)

	t.Logf("PlatformInformation(): %v, %v, %v", platform, family, version)
}

func BenchmarkBootTimeWithCache(b *testing.B) {
	EnableBootTimeCache(true)
	for i := 0; i < b.N; i++ {
		BootTime()
	}
}

func BenchmarkBootTimeWithoutCache(b *testing.B) {
	EnableBootTimeCache(false)
	for i := 0; i < b.N; i++ {
		BootTime()
	}
}
