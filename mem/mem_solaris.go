package mem

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/internal/common"
)

type VirtualMemoryZoneStat struct {
	Name             string        `json:"zonename"`
	AnonAllocFail    int64         `json:"anon_alloc_fail"`    // cnt of anon alloc fails
	AnonPageIn       int64         `json:"anonpgin"`           // anon pages paged in
	Class            string        `json:"class"`              //
	CreateTime       float64       `json:"crtime"`             // creation time (from gethrtime())
	ExecPageIn       int64         `json:"execpgin"`           // exec pages paged in
	FilesystemPageIn int64         `json:"fspgin"`             // fs pages paged in
	NPFThrottle      int64         `json:"n_pf_throttle"`      // cnt of page flt throttles
	NPFThrottleTime  time.Duration `json:"n_pf_throttle_usec"` // time of page flt throttles
	NumOver          int64         `json:"nover"`              // # of times over phys. cap
	PagedOut         int64         `json:"pagedout"`           // bytes of mem. paged out
	PagesPagedIn     int64         `json:"pgpgin"`             // pages paged in
	PhysicalCap      int64         `json:"physcap"`            // current phys. memory limit
	RSS              int64         `json:"rss"`                // current bytes of phys. mem. (RSS)
	Swap             int64         `json:"swap"`               // bytes of swap reserved by zone
	SwapCap          int64         `json:"swapcap"`            // current swap limit
}

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
	type _vmZoneStat struct {
		VirtualMemoryZoneStat
		NPFThrottleTime int64 `json:"n_pf_throttle_usec"`
	}
	type _kstatFrame struct {
		Instance int         `json:"instance"`
		Data     _vmZoneStat `json:"data"`
	}
	kstatFrames := make([]_kstatFrame, 0, 1)
	err = json.Unmarshal(kstatJSON, &kstatFrames)
	if err != nil {
		return nil, err
	}

	ret := &VirtualMemoryZoneStat{}
	for _, zoneStat := range kstatFrames {
		ret = &zoneStat.Data.VirtualMemoryZoneStat
		ret.NPFThrottleTime = time.Duration(zoneStat.Data.NPFThrottleTime) * time.Microsecond

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
