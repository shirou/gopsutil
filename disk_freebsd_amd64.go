// +build freebsd
// +build amd64

package gopsutil

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
