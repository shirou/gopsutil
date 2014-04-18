package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDisk_usage(t *testing.T) {
	disk := NewDisk()

	v, err := disk.Disk_usage("/")
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}
