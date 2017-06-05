// +build solaris

package mem

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/mailru/easyjson"
	"github.com/shirou/gopsutil/internal/common"
)

// VirtualMemoryZone returns the memory used inside of the visible zone(s).
func VirtualMemoryZone() (*VirtualMemoryZoneStat, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return nil, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	kstatJSON, err := invoke.Command(kstatPath, "-jm", "memory_cap")
	if err != nil {
		return nil, fmt.Errorf("cannot execute kstat(1M): %v", err)
	}
	kstatFrames := _memoryCapKStatFrames{}
	if err := easyjson.Unmarshal(kstatJSON, &kstatFrames); err != nil {
		return nil, fmt.Errorf("cannot decode kstat(1M) memory_cap JSON: %v", err)
	}

	var ret *VirtualMemoryZoneStat
	for _, zoneStat := range kstatFrames {
		ret = &zoneStat.Data
		// Fixup NPFThrottleTime: kstat(1M) reports this value in µs but the value
		// was Unmarshaled as ns
		ret.NPFThrottleTime = ret.NPFThrottleTime * time.Microsecond

		// XXX Only one object for now: zones can't be nested
		break
	}

	return ret, nil
}

// VirtualMemory for Solaris is a minimal implementation which only returns
// what Nomad needs. It does take into account global vs zone, however.
func VirtualMemory() (*VirtualMemoryStat, error) {
	inZone, err := common.InZone()
	if err != nil {
		return nil, err
	}

	if inZone {
		vmStat, err := VirtualMemoryZone()
		if err != nil {
			return nil, err
		}

		// FIXME(seanc@): I'm unsure as to how to square all of the parameters here
		// with what's available in Linux.  It's not that Illumos needs this data,
		// it's that these counters aren't exposed to the individual zone itself.
		// https://github.com/torvalds/linux/blob/master/Documentation/filesystems/proc.txt#L874
		usedPct := float64(vmStat.RSS) / float64(vmStat.PhysicalCap)
		if math.IsNaN(usedPct) {
			usedPct = 0.0 // UNLIMITED BITS!!!
		} else {
			usedPct = usedPct * 100.0
		}

		ret := &VirtualMemoryStat{
			Total:        uint64(vmStat.PhysicalCap),
			Available:    uint64(vmStat.PhysicalCap - vmStat.RSS),
			Used:         uint64(vmStat.RSS),
			UsedPercent:  usedPct,
			Free:         uint64(vmStat.PhysicalCap - vmStat.RSS - 1),
			Active:       uint64(vmStat.RSS),
			Inactive:     uint64(vmStat.PhysicalCap - vmStat.RSS),
			Wired:        0,
			Buffers:      0,
			Cached:       0,
			Writeback:    0,
			Dirty:        0,
			WritebackTmp: 0,
			Shared:       0,
			Slab:         0,
			PageTables:   0,
			SwapCached:   0,
		}

		return ret, nil
	}

	cap, err := nonGlobalZoneMemoryCapacity()
	if err != nil {
		return nil, err
	}
	ret := &VirtualMemoryStat{
		Total: cap,
	}

	return ret, nil
}

func SwapMemory() (*SwapMemoryStat, error) {
	inZone, err := common.InZone()
	if err != nil {
		return nil, err
	}

	if inZone {
		vmStat, err := VirtualMemoryZone()
		if err != nil {
			return nil, err
		}

		usedPct := float64(vmStat.Swap) / float64(vmStat.SwapCap)
		if math.IsNaN(usedPct) {
			usedPct = 0.0
		} else {
			usedPct = usedPct * 100.0
		}

		ret := &SwapMemoryStat{
			Total:       uint64(vmStat.SwapCap),
			Used:        uint64(vmStat.Swap),
			Free:        uint64(vmStat.SwapCap - vmStat.Swap),
			UsedPercent: usedPct,
			// FIXME(seanc@): Number of pages swapped in/out since the last
			// measurement.
			Sin:  0.0,
			Sout: 0.0,
		}
		return ret, nil
	}

	return nil, common.ErrNotImplementedError
}

var kstatMatch = regexp.MustCompile(`([^\s]+)[\s]+([^\s]*)`)

func nonGlobalZoneMemoryCapacity() (uint64, error) {
	kstatPath, err := common.KStatPath()
	if err != nil {
		return 0, fmt.Errorf("cannot find kstat(1M): %v", err)
	}

	out, err := invoke.Command(kstatPath, "-p", "-c", "zone_memory_cap", "memory_cap:*:*:physcap")
	if err != nil {
		return 0, err
	}

	kstats := kstatMatch.FindAllStringSubmatch(string(out), -1)
	if len(kstats) != 1 {
		return 0, fmt.Errorf("expected 1 kstat, found %d", len(kstats))
	}

	memSizeBytes, err := strconv.ParseUint(kstats[0][2], 10, 64)
	if err != nil {
		return 0, err
	}

	return memSizeBytes, nil
}
