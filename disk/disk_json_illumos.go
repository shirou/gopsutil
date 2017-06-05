// +build solaris

package disk

import "time"

//go:generate easyjson -output_filename disk_json_illumos.generated.go disk_json_illumos.go

// IOCountersZoneStat measurements are defined in
// https://github.com/joyent/illumos-joyent/blob/master/usr/src/uts/common/sys/zone.h#L404-L422
// however the actual members are detailed elsewhere in kstat.h:
// https://github.com/illumos/illumos-gate/blob/master/usr/src/uts/common/sys/kstat.h#L588-L685
type IOCountersZoneStat struct {
	Name string `json:"zonename"`

	Num10msOps  uint64 `json:"10ms_ops"`  // number of IOs >=10ms
	Num100msOps uint64 `json:"100ms_ops"` // number of IOs >= 100ms
	Num1sOps    uint64 `json:"1s_ops"`    // number of IOs >= 1s
	Num10sOps   uint64 `json:"10s_ops"`   // number of IOs >= 10s

	// Delay* is the cumulative number of operations and time that the ZFS IO
	// scheduler spent throttling IO for a given zone (see zone_io_delay):
	// https://github.com/joyent/illumos-joyent/blob/master/usr/src/uts/common/fs/zfs/zfs_zone.c#L954-L995
	// https://github.com/joyent/illumos-joyent/blob/master/usr/src/uts/common/fs/zfs/zfs_zone.c#L1046-L1063
	DelayCount uint64        `json:"delay_cnt"`  // number of times IO has been delayed by ZFS IO scheduler
	DelayTime  time.Duration `json:"delay_time"` // cumulative wait time IO throttled by ZFS IO scheduler

	BytesRead    uint64 `json:"nread"`    // number of bytes read
	BytesWritten uint64 `json:"nwritten"` // number of bytes written
	NumReadOps   uint64 `json:"reads"`    // number of read operations
	NumWriteOps  uint64 `json:"writes"`   // number of write operations

	NumRunningIO  uint64        `json:"rcnt"`  // count of elements in run state
	TimeIORunning time.Duration `json:"rtime"` // cumulative run (service) time

	NumQueuedIO  uint64        `json:"wcnt"`  // count of elements in wait state
	TimeIOQueued time.Duration `json:"wtime"` // cumulative wait (pre-service) time

	// See
	// github.com/illumos/illumos-gate/blob/master/usr/src/uts/common/sys/kstat.h#L588-L685
	// for an explanation of these stats
	RiemannSumRunning uint64 `json:"rlentime"` // cumulative run length*time product
	RiemannSumQueued  uint64 `json:"wlentime"` // cumulative wait length*time product
}

//easyjson:json
type _zoneVFSKStatFrame struct {
	Instance int                `json:"instance"`
	Data     IOCountersZoneStat `json:"data"`
}

//easyjson:json
type _zoneVFSKStatFrames []_zoneVFSKStatFrame
