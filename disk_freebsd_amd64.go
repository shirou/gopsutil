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
	FVersion     uint32           /* structure version number */
	FType        uint32           /* type of filesystem */
	FFlags       uint64           /* copy of mount exported flags */
	FBsize       uint64           /* filesystem fragment size */
	FIosize      uint64           /* optimal transfer block size */
	FBlocks      uint64           /* total data blocks in filesystem */
	FBfree       uint64           /* free blocks in filesystem */
	FBavail      int64            /* free blocks avail to non-superuser */
	FFiles       uint64           /* total file nodes in filesystem */
	FFfree       int64            /* free nodes avail to non-superuser */
	FSyncwrites  uint64           /* count of sync writes since mount */
	FAsyncwrites uint64           /* count of async writes since mount */
	FSyncreads   uint64           /* count of sync reads since mount */
	FAsyncreads  uint64           /* count of async reads since mount */
	FSpare       [10]uint64       /* unused spare */
	FNamemax     uint32           /* maximum filename length */
	FOwner       uint32           /* user that mounted the filesystem */
	FFsid        int32            /* filesystem id */
	FCharspare   [80]byte         /* spare string space */
	FFstypename  [MFSNAMELEN]byte /* filesystem type name */
	FMntfromname [MNAMELEN]byte   /* mounted filesystem */
	FMntonname   [MNAMELEN]byte   /* directory on which mounted */
}

// /usr/include/devstat.h
// devstat_getdevs()
// kern.devstat.all -> devstats list struct

// struct devinfo {
//         struct devstat  *devices;
//         u_int8_t        *mem_ptr;
//         long            generation;
//         int             numdevs;
// };
//
// struct statinfo {
//         long            cp_time[CPUSTATES];
//         long            tk_nin;
//         long            tk_nout;
//         struct devinfo  *dinfo;
//         long double     snap_time;
// };

// /usr/include/devinfo.h

//    struct devinfo_dev {
//               devinfo_handle_t    dd_handle;      /* device handle */
//               devinfo_handle_t    dd_parent;      /* parent handle */
//               char                *dd_name;       /* name of device */
//               char                *dd_desc;       /* device description */
//               char                *dd_drivername; /* name of attached driver */
//               char                *dd_pnpinfo;    /* pnp info from parent bus */
//               char                *dd_location;   /* Where bus thinks dev at */
//               uint32_t            dd_devflags;    /* API flags */
//               uint16_t            dd_flags;       /* internal dev flags */
//               device_state_t      dd_state;       /* attachment state of dev */
//	};
//
//           struct devinfo_rman {
//               devinfo_handle_t    dm_handle;      /* resource manager handle */
//               u_long              dm_start;       /* resource start */
//               u_long              dm_size;        /* resource size */
//               char                *dm_desc;       /* resource description */
//           };
//
//           struct devinfo_res {
//               devinfo_handle_t    dr_handle;      /* resource handle */
//               devinfo_handle_t    dr_rman;        /* resource manager handle */
//               devinfo_handle_t    dr_device;      /* owning device */
//               u_long              dr_start;       /* region start */
//               u_long              dr_size;        /* region size */
//           };
