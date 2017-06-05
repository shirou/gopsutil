// +build solaris

package mem

//go:generate easyjson -output_filename mem_json_illumos.generated.go mem_json_illumos.go

import "time"

//easyjson:json
type VirtualMemoryZoneStat struct {
	Name             string        `json:"zonename"`
	AnonAllocFail    uint64        `json:"anon_alloc_fail"`    // cnt of anon alloc fails
	AnonPageIn       uint64        `json:"anonpgin"`           // anon pages paged in
	Class            string        `json:"class"`              //
	ExecPageIn       uint64        `json:"execpgin"`           // exec pages paged in
	FilesystemPageIn uint64        `json:"fspgin"`             // fs pages paged in
	NPFThrottle      uint64        `json:"n_pf_throttle"`      // cnt of page flt throttles
	NPFThrottleTime  time.Duration `json:"n_pf_throttle_usec"` // time of page flt throttles
	NumOver          uint64        `json:"nover"`              // # of times over phys. cap
	PagedOut         uint64        `json:"pagedout"`           // bytes of mem. paged out
	PagesPagedIn     uint64        `json:"pgpgin"`             // pages paged in
	PhysicalCap      uint64        `json:"physcap"`            // current phys. memory limit
	RSS              uint64        `json:"rss"`                // current bytes of phys. mem. (RSS)
	Swap             uint64        `json:"swap"`               // bytes of swap reserved by zone
	SwapCap          uint64        `json:"swapcap"`            // current swap limit
}

//easyjson:json
type _memoryCapKStatFrame struct {
	Instance int                   `json:"instance"`
	Data     VirtualMemoryZoneStat `json:"data"`
}

//easyjson:json
type _memoryCapKStatFrames []_memoryCapKStatFrame
