// SPDX-License-Identifier: BSD-3-Clause
package disk

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/internal/common"
)

func TestUsage(t *testing.T) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	v, err := Usage(path)
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.Equalf(t, v.Path, path, "error %v", err)
}

func TestPartitions(t *testing.T) {
	ret, err := Partitions(false)
	common.SkipIfNotImplementedErr(t, err)
	if err != nil || len(ret) == 0 {
		t.Errorf("error %v", err)
	}
	t.Log(ret)

	assert.NotEmptyf(t, ret, "ret is empty")
	for _, disk := range ret {
		assert.NotEmptyf(t, disk.Device, "Could not get device info %v", disk)
	}
}

func TestIOCounters(t *testing.T) {
	ret, err := IOCounters()
	common.SkipIfNotImplementedErr(t, err)
	require.NoError(t, err)
	assert.NotEmptyf(t, ret, "ret is empty")
	empty := IOCountersStat{}
	for part, io := range ret {
		t.Log(part, io)
		assert.NotEqualf(t, io, empty, "io_counter error %v, %v", part, io)
	}
}

// https://github.com/shirou/gopsutil/issues/560 regression test
func TestIOCounters_concurrency_on_darwin_cgo(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	var wg sync.WaitGroup
	const maxCount = 1000
	for i := 1; i < maxCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			IOCounters()
		}()
	}
	wg.Wait()
}

func TestUsageStat_String(t *testing.T) {
	v := UsageStat{
		Path:              "/",
		Total:             1000,
		Free:              2000,
		Used:              3000,
		UsedPercent:       50.1,
		InodesTotal:       4000,
		InodesUsed:        5000,
		InodesFree:        6000,
		InodesUsedPercent: 49.1,
		Fstype:            "ext4",
	}
	e := `{"path":"/","fstype":"ext4","total":1000,"free":2000,"used":3000,"usedPercent":50.1,"inodesTotal":4000,"inodesUsed":5000,"inodesFree":6000,"inodesUsedPercent":49.1}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "DiskUsageStat string is invalid: %v", v)
}

func TestPartitionStat_String(t *testing.T) {
	v := PartitionStat{
		Device:     "sd01",
		Mountpoint: "/",
		Fstype:     "ext4",
		Opts:       []string{"ro"},
	}
	e := `{"device":"sd01","mountpoint":"/","fstype":"ext4","opts":["ro"]}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "DiskUsageStat string is invalid: %v", v)
}

func TestIOCountersStat_String(t *testing.T) {
	v := IOCountersStat{
		Name:         "sd01",
		ReadCount:    100,
		WriteCount:   200,
		ReadBytes:    300,
		WriteBytes:   400,
		SerialNumber: "SERIAL",
	}
	e := `{"readCount":100,"mergedReadCount":0,"writeCount":200,"mergedWriteCount":0,"readBytes":300,"writeBytes":400,"readTime":0,"writeTime":0,"iopsInProgress":0,"ioTime":0,"weightedIO":0,"name":"sd01","serialNumber":"SERIAL","label":""}`
	assert.JSONEqf(t, e, fmt.Sprintf("%v", v), "DiskUsageStat string is invalid: %v", v)
}
