// +build solaris

package disk

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/internal/common"
	"golang.org/x/sys/unix"
)

const (
	// _DEFAULT_NUM_MOUNTS is set to `cat /etc/mnttab | wc -l` rounded up to the
	// nearest power of two.
	_DEFAULT_NUM_MOUNTS = 32

	// _MNTTAB default place to read mount information
	_MNTTAB = "/etc/mnttab"
)

var (
	// A blacklist of read-only virtual filesystems.  Writable filesystems are of
	// operational concern and must not be included in this list.
	fsTypeBlacklist = map[string]struct{}{
		"ctfs":   struct{}{},
		"dev":    struct{}{},
		"fd":     struct{}{},
		"lofs":   struct{}{},
		"lxproc": struct{}{},
		"mntfs":  struct{}{},
		"objfs":  struct{}{},
		"proc":   struct{}{},
	}
)

// IOCountersZoneStat measurements are defined in
// https://github.com/joyent/illumos-joyent/blob/master/usr/src/uts/common/sys/zone.h#L404-L422
// however the actual members are detailed elsewhere in kstat.h:
// https://github.com/illumos/illumos-gate/blob/master/usr/src/uts/common/sys/kstat.h#L588-L685
//
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

func Partitions(all bool) ([]PartitionStat, error) {
	ret := make([]PartitionStat, 0, _DEFAULT_NUM_MOUNTS)

	// Scan mnttab(4)
	f, err := os.Open(_MNTTAB)
	if err != nil {
	}
	defer func() {
		if err == nil {
			err = f.Close()
		} else {
			f.Close()
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")

		if _, found := fsTypeBlacklist[fields[2]]; found {
			continue
		}

		ret = append(ret, PartitionStat{
			// NOTE(seanc@): Device isn't exactly accurate: from mnttab(4): "The name
			// of the resource that has been mounted."  Ideally this value would come
			// from Statvfs_t.Fsid but I'm leaving it to the caller to traverse
			// unix.Statvfs().
			Device:     fields[0],
			Mountpoint: fields[1],
			Fstype:     fields[2],
			Opts:       fields[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to scan %q: %v", _MNTTAB, err)
	}

	return ret, err
}

func IOCounters(names ...string) (map[string]IOCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func Usage(path string) (*UsageStat, error) {
	statvfs := unix.Statvfs_t{}
	if err := unix.Statvfs(path, &statvfs); err != nil {
		return nil, fmt.Errorf("unable to call statvfs(2) on %q: %v", path, err)
	}

	usageStat := &UsageStat{
		Path:   path,
		Fstype: common.IntToString(statvfs.Basetype[:]),
		Total:  statvfs.Blocks * statvfs.Frsize,
		Free:   statvfs.Bfree * statvfs.Frsize,
		Used:   (statvfs.Blocks - statvfs.Bfree) * statvfs.Frsize,

		// NOTE: ZFS (and FreeBZSD's UFS2) use dynamic inode/dnode allocation.
		// Explicitly return a near-zero value for InodesUsedPercent so that nothing
		// attempts to garbage collect based on a lack of available inodes/dnodes.
		// Similarly, don't use the zero value to prevent divide-by-zero situations
		// and inject a faux near-zero value.  Filesystems evolve.  Has your
		// filesystem evolved?  Probably not if you care about the number of
		// available inodes.
		InodesTotal:       1024.0 * 1024.0,
		InodesUsed:        1024.0,
		InodesFree:        math.MaxUint64,
		InodesUsedPercent: (1024.0 / (1024.0 * 1024.0)) * 100.0,
	}

	usageStat.UsedPercent = (float64(usageStat.Used) / float64(usageStat.Total)) * 100.0

	return usageStat, nil
}
