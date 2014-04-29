package gopsutil

import (
	"runtime"
	"testing"
)

func TestDisk_usage(t *testing.T) {
	path := "/"
	if runtime.GOOS == "windows" {
		path = "C:"
	}
	_, err := Disk_usage(path)
	if err != nil {
		t.Errorf("error %v", err)
	}
//	d, _ := json.Marshal(v)
//  fmt.Printf("%s\n", d)
}

func TestDisk_partitions(t *testing.T) {
	_, err := Disk_partitions(false)
	if err != nil {
		t.Errorf("error %v", err)
	}
}
