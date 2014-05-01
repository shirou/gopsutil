// +build freebsd

package gopsutil

import (
	"errors"
	"syscall"
	"unsafe"
)

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat

	// get length
	count, err := syscall.Getfsstat(nil, MNT_WAIT)
	if err != nil {
		return ret, err
	}

	fs := make([]Statfs, 0, count)
	_, err = Getfsstat(fs, MNT_WAIT)

	for _, stat := range fs {
		opts := "rw"
		if stat.FFlags&MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.FFlags&MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.FFlags&MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.FFlags&MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.FFlags&MNT_UNION != 0 {
			opts += ",union"
		}
		if stat.FFlags&MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.FFlags&MNT_SUIDDIR != 0 {
			opts += ",suiddir"
		}
		if stat.FFlags&MNT_SOFTDEP != 0 {
			opts += ",softdep"
		}
		if stat.FFlags&MNT_NOSYMFOLLOW != 0 {
			opts += ",nosymfollow"
		}
		if stat.FFlags&MNT_GJOURNAL != 0 {
			opts += ",gjounalc"
		}
		if stat.FFlags&MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if stat.FFlags&MNT_ACLS != 0 {
			opts += ",acls"
		}
		if stat.FFlags&MNT_NOATIME != 0 {
			opts += ",noattime"
		}
		if stat.FFlags&MNT_NOCLUSTERR != 0 {
			opts += ",nocluster"
		}
		if stat.FFlags&MNT_NOCLUSTERW != 0 {
			opts += ",noclusterw"
		}
		if stat.FFlags&MNT_NFS4ACLS != 0 {
			opts += ",nfs4acls"
		}

		d := DiskPartitionStat{
			Mountpoint: byteToString(stat.FMntonname[:]),
			Fstype:     byteToString(stat.FFstypename[:]),
			Opts:       opts,
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	ret := make(map[string]DiskIOCountersStat, 0)
	return ret, errors.New("not implemented yet")
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
