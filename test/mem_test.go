package test

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil"
)

func TestVirtual_memory(t *testing.T) {
	v, err := gopsutil.VirtualMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}

	empty := &gopsutil.VirtualMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
}

func TestSwap_memory(t *testing.T) {
	v, err := gopsutil.SwapMemory()
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &gopsutil.SwapMemoryStat{}
	if v == empty {
		t.Errorf("error %v", v)
	}
}

func TestVirtualMemoryStat_String(t *testing.T) {
	v := gopsutil.VirtualMemoryStat{
		Total:       10,
		Available:   20,
		Used:        30,
		UsedPercent: 30.1,
		Free:        40,
	}
	e := `{"total":10,"available":20,"used":30,"usedPercent":30.1,"free":40,"active":0,"inactive":0,"buffers":0,"cached":0,"wired":0,"shared":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("VirtualMemoryStat string is invalid: %v", v)
	}
}

func TestSwapMemoryStat_String(t *testing.T) {
	v := gopsutil.SwapMemoryStat{
		Total:       10,
		Used:        30,
		Free:        40,
		UsedPercent: 30.1,
	}
	e := `{"total":10,"used":30,"free":40,"usedPercent":30.1,"sin":0,"sout":0}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("SwapMemoryStat string is invalid: %v", v)
	}
}
