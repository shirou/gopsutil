package gopsutil

import (
	"fmt"
	"runtime"
	"testing"
)

func TestDisk_usage(t *testing.T) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	v, err := DiskUsage(path)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v.Path != path {
		t.Errorf("error %v", err)
	}
}

func TestDisk_partitions(t *testing.T) {
	ret, err := DiskPartitions(false)
	if err != nil || len(ret) == 0 {
		t.Errorf("error %v", err)
	}
	empty := DiskPartitionStat{}
	for _, disk := range ret {
		if disk == empty {
			t.Errorf("Could not get device info %v", disk)
		}
	}
}

func TestDisk_io_counters(t *testing.T) {
	ret, err := DiskIOCounters()
	if err != nil || len(ret) == 0 {
		t.Errorf("error %v", err)
	}
	empty := DiskIOCountersStat{}
	for part, io := range ret {
		if io == empty {
			t.Errorf("io_counter error %v, %v", part, io)
		}
	}
}

func TestDiskUsageStat_String(t *testing.T) {
	v := DiskUsageStat{
		Path:        "/",
		Total:       1000,
		Free:        2000,
		Used:        3000,
		UsedPercent: 50.1,
		InodesTotal: 10,
		InodesUsed:  1,
		InodesFree:  9,
		InodesUsedPercent: 1.,
	}
	e := `{"path":"/","total":1000,"free":2000,"used":3000,"usedPercent":50.1,"inodesTotal":10,"inodesUsed":1,"inodesFree":9,"inodesUsedPercent":1}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}

func TestDiskPartitionStat_String(t *testing.T) {
	v := DiskPartitionStat{
		Device:     "sd01",
		Mountpoint: "/",
		Fstype:     "ext4",
		Opts:       "ro",
	}
	e := `{"device":"sd01","mountpoint":"/","fstype":"ext4","opts":"ro"}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}

func TestDiskIOCountersStat_String(t *testing.T) {
	v := DiskIOCountersStat{
		Name:       "sd01",
		ReadCount:  100,
		WriteCount: 200,
		ReadBytes:  300,
		WriteBytes: 400,
	}
	e := `{"readCount":100,"writeCount":200,"readBytes":300,"writeBytes":400,"readTime":0,"writeTime":0,"name":"sd01","ioTime":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("DiskUsageStat string is invalid: %v", v)
	}
}
