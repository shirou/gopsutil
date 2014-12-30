// +build freebsd

package disk

import (
	"syscall"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
)

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat

	// get length
	count, err := syscall.Getfsstat(nil, MntWait)
	if err != nil {
		return ret, err
	}

	fs := make([]Statfs, count)
	_, err = Getfsstat(fs, MntWait)

	for _, stat := range fs {
		opts := "rw"
		if stat.FFlags&MntReadOnly != 0 {
			opts = "ro"
		}
		if stat.FFlags&MntSynchronous != 0 {
			opts += ",sync"
		}
		if stat.FFlags&MntNoExec != 0 {
			opts += ",noexec"
		}
		if stat.FFlags&MntNoSuid != 0 {
			opts += ",nosuid"
		}
		if stat.FFlags&MntUnion != 0 {
			opts += ",union"
		}
		if stat.FFlags&MntAsync != 0 {
			opts += ",async"
		}
		if stat.FFlags&MntSuidDir != 0 {
			opts += ",suiddir"
		}
		if stat.FFlags&MntSoftDep != 0 {
			opts += ",softdep"
		}
		if stat.FFlags&MntNoSymFollow != 0 {
			opts += ",nosymfollow"
		}
		if stat.FFlags&MntGEOMJournal != 0 {
			opts += ",gjounalc"
		}
		if stat.FFlags&MntMultilabel != 0 {
			opts += ",multilabel"
		}
		if stat.FFlags&MntACLs != 0 {
			opts += ",acls"
		}
		if stat.FFlags&MntNoATime != 0 {
			opts += ",noattime"
		}
		if stat.FFlags&MntClusterRead != 0 {
			opts += ",nocluster"
		}
		if stat.FFlags&MntClusterWrite != 0 {
			opts += ",noclusterw"
		}
		if stat.FFlags&MntNFS4ACLs != 0 {
			opts += ",nfs4acls"
		}

		d := DiskPartitionStat{
			Device:     common.ByteToString(stat.FMntfromname[:]),
			Mountpoint: common.ByteToString(stat.FMntonname[:]),
			Fstype:     common.ByteToString(stat.FFstypename[:]),
			Opts:       opts,
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	return nil, common.NotImplementedError

	// statinfo->devinfo->devstat
	// /usr/include/devinfo.h

	// get length
	count, err := Getfsstat(nil, MntWait)
	if err != nil {
		return nil, err
	}

	fs := make([]Statfs, count)
	_, err = Getfsstat(fs, MntWait)

	ret := make(map[string]DiskIOCountersStat, 0)
	for _, stat := range fs {
		name := common.ByteToString(stat.FMntonname[:])
		d := DiskIOCountersStat{
			Name:       name,
			ReadCount:  stat.FSyncwrites + stat.FAsyncwrites,
			WriteCount: stat.FSyncreads + stat.FAsyncreads,
		}

		ret[name] = d
	}

	return ret, nil
}

// Getfsstat is borrowed from pkg/syscall/syscall_freebsd.go
// change Statfs_t to Statfs in order to get more information
func Getfsstat(buf []Statfs, flags int) (n int, err error) {
	var _p0 unsafe.Pointer
	var bufsize uintptr
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
		bufsize = unsafe.Sizeof(Statfs{}) * uintptr(len(buf))
	}
	r0, _, e1 := syscall.Syscall(syscall.SYS_GETFSSTAT, uintptr(_p0), bufsize, uintptr(flags))
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}
