package gopsutil

import (
	"fmt"
	"testing"
)

func TestVirtual_memory(t *testing.T) {
	v, err := VirtualMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	fmt.Println(v)
}

func TestSwap_memory(t *testing.T) {
	v, err := SwapMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	fmt.Println(v)
}
