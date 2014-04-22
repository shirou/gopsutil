package gopsutil

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
)

func TestDisk_usage(t *testing.T) {
	disk := NewDisk()

	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	v, err := disk.Disk_usage(path)
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}

func TestDisk_partitions(t *testing.T) {
	disk := NewDisk()

	v, err := disk.Disk_partitions()
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}
