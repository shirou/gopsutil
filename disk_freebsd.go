// +build freebsd

package gopsutil

import (
	"syscall"
	"unsafe"
)

const (
	MNT_WAIT   = 1
	MFSNAMELEN = 16 /* length of type name including null */
	MNAMELEN   = 88 /* size of on/from name bufs */
)

// sys/mount.h
const (
	MNT_RDONLY      = 0x00000001 /* read only filesystem */
	MNT_SYNCHRONOUS = 0x00000002 /* filesystem written synchronously */
	MNT_NOEXEC      = 0x00000004 /* can't exec from filesystem */
	MNT_NOSUID      = 0x00000008 /* don't honor setuid bits on fs */
	MNT_UNION       = 0x00000020 /* union with underlying filesystem */
	MNT_ASYNC       = 0x00000040 /* filesystem written asynchronously */
	MNT_SUIDDIR     = 0x00100000 /* special handling of SUID on dirs */
	MNT_SOFTDEP     = 0x00200000 /* soft updates being done */
	MNT_NOSYMFOLLOW = 0x00400000 /* do not follow symlinks */
	MNT_GJOURNAL    = 0x02000000 /* GEOM journal support enabled */
	MNT_MULTILABEL  = 0x04000000 /* MAC support for individual objects */
	MNT_ACLS        = 0x08000000 /* ACL support enabled */
	MNT_NOATIME     = 0x10000000 /* disable update of file access time */
	MNT_NOCLUSTERR  = 0x40000000 /* disable cluster read */
	MNT_NOCLUSTERW  = 0x80000000 /* disable cluster write */
	MNT_NFS4ACLS    = 0x00000010
)

type Statfs struct {
	F_version     uint32           /* structure version number */
	F_type        uint32           /* type of filesystem */
	F_flags       uint64           /* copy of mount exported flags */
	F_bsize       uint64           /* filesystem fragment size */
	F_iosize      uint64           /* optimal transfer block size */
	F_blocks      uint64           /* total data blocks in filesystem */
	F_bfree       uint64           /* free blocks in filesystem */
	F_bavail      int64            /* free blocks avail to non-superuser */
	F_files       uint64           /* total file nodes in filesystem */
	F_ffree       int64            /* free nodes avail to non-superuser */
	F_syncwrites  uint64           /* count of sync writes since mount */
	F_asyncwrites uint64           /* count of async writes since mount */
	F_syncreads   uint64           /* count of sync reads since mount */
	F_asyncreads  uint64           /* count of async reads since mount */
	F_spare       [10]uint64       /* unused spare */
	F_namemax     uint32           /* maximum filename length */
	F_owner       uint32           /* user that mounted the filesystem */
	F_fsid        int32            /* filesystem id */
	F_charspare   [80]byte         /* spare string space */
	F_fstypename  [MFSNAMELEN]byte /* filesystem type name */
	F_mntfromname [MNAMELEN]byte   /* mounted filesystem */
	F_mntonname   [MNAMELEN]byte   /* directory on which mounted */
}

func Disk_partitions() ([]Disk_partitionStat, error) {
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
