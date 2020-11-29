package mem

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVirtualMemoryEx(t *testing.T) {
	v, err := VirtualMemoryEx()
	if err != nil {
		t.Error(err)
	}

	t.Log(v)
}

var virtualMemoryTests = []struct {
	mockedRootFS string
	stat         *VirtualMemoryStat
}{
	{"intelcorei5", &VirtualMemoryStat{
		Total:          16502300672,
		Available:      11495358464,
		Used:           3437277184,
		UsedPercent:    20.82907863769651,
		Free:           8783491072,
		Active:         4347392000,
		Inactive:       2938834944,
		Wired:          0,
		Laundry:        0,
		Buffers:        212496384,
		Cached:         4069036032,
		Writeback:      0,
		Dirty:          176128,
		WritebackTmp:   0,
		Shared:         1222402048,
		Slab:           253771776,
		SReclaimable:   186470400,
		SUnreclaim:     67301376,
		PageTables:     65241088,
		SwapCached:     0,
		CommitLimit:    16509730816,
		CommittedAS:    12360818688,
		HighTotal:      0,
		HighFree:       0,
		LowTotal:       0,
		LowFree:        0,
		SwapTotal:      8258580480,
		SwapFree:       8258580480,
		Mapped:         1172627456,
		VMallocTotal:   35184372087808,
		VMallocUsed:    0,
		VMallocChunk:   0,
		HugePagesTotal: 0,
		HugePagesFree:  0,
		HugePageSize:   2097152},
	},
}

func TestVirtualMemoryLinux(t *testing.T) {
	origProc := os.Getenv("HOST_PROC")
	defer os.Setenv("HOST_PROC", origProc)

	for _, tt := range virtualMemoryTests {
		t.Run(tt.mockedRootFS, func(t *testing.T) {
			os.Setenv("HOST_PROC", filepath.Join("testdata/linux/virtualmemory/", tt.mockedRootFS, "proc"))

			stat, err := VirtualMemory()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if !reflect.DeepEqual(stat, tt.stat) {
				t.Errorf("got: %+v\nwant: %+v", stat, tt.stat)
			}
		})
	}
}
