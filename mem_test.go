package gopsutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestVirtual_memory(t *testing.T) {
	mem := NewMem()

	v, err := mem.Virtual_memory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}

func TestSwap_memory(t *testing.T) {
	mem := NewMem()

	v, err := mem.Swap_memory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	d, _ := json.Marshal(v)
	fmt.Printf("%s\n", d)
}
