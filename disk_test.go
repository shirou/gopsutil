package gopsutil

import (
	//	"fmt"
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

func TestDisk_io_counters(t *testing.T) {
	ret, err := Disk_io_counters()
	if err != nil {
		t.Errorf("error %v", err)
	}
	for _, io := range ret {
		if io.Name == "" {
			t.Errorf("io_counter error %v", io)
		}
	}
}
