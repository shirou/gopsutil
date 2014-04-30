package gopsutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestVirtual_memory(t *testing.T) {
	v, err := VirtualMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}

func TestSwap_memory(t *testing.T) {
	v, err := SwapMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}
