// +build freebsd

package gopsutil

import (
	"syscall"
	"unsafe"
)

func Disk_partitions(all bool) ([]Disk_partitionStat, error) {
	ret := make([]Disk_partitionStat, 0)

	// get length
	count, err := syscall.Getfsstat(nil, MNT_WAIT)
	if err != nil {
		return ret, err
	}

	fs := make([]Statfs, count)
	_, err = Getfsstat(fs, MNT_WAIT)

	for _, stat := range fs {
		opts := "rw"
		if stat.F_flags&MNT_RDONLY != 0 {
			opts = "ro"
		}
		if stat.F_flags&MNT_SYNCHRONOUS != 0 {
			opts += ",sync"
		}
		if stat.F_flags&MNT_NOEXEC != 0 {
			opts += ",noexec"
		}
		if stat.F_flags&MNT_NOSUID != 0 {
			opts += ",nosuid"
		}
		if stat.F_flags&MNT_UNION != 0 {
			opts += ",union"
		}
		if stat.F_flags&MNT_ASYNC != 0 {
			opts += ",async"
		}
		if stat.F_flags&MNT_SUIDDIR != 0 {
			opts += ",suiddir"
		}
		if stat.F_flags&MNT_SOFTDEP != 0 {
			opts += ",softdep"
		}
		if stat.F_flags&MNT_NOSYMFOLLOW != 0 {
			opts += ",nosymfollow"
		}
		if stat.F_flags&MNT_GJOURNAL != 0 {
			opts += ",gjounalc"
		}
		if stat.F_flags&MNT_MULTILABEL != 0 {
			opts += ",multilabel"
		}
		if stat.F_flags&MNT_ACLS != 0 {
			opts += ",acls"
		}
		if stat.F_flags&MNT_NOATIME != 0 {
			opts += ",noattime"
		}
		if stat.F_flags&MNT_NOCLUSTERR != 0 {
			opts += ",nocluster"
		}
		if stat.F_flags&MNT_NOCLUSTERW != 0 {
			opts += ",noclusterw"
		}
		if stat.F_flags&MNT_NFS4ACLS != 0 {
			opts += ",nfs4acls"
		}

		d := Disk_partitionStat{
			Mountpoint: byteToString(stat.F_mntonname[:]),
			Fstype:     byteToString(stat.F_fstypename[:]),
			Opts:       opts,
		}
		ret = append(ret, d)
	}

	return ret, nil
}

// This is borrowed from pkg/syscall/syscall_freebsd.go
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
