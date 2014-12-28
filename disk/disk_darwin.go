// +build darwin

package gopsutil

import (
	common "github.com/shirou/gopsutil/common"

	"golang.org/x/sys/unix"
)

// sys/mount.h
const (
	MntReadOnly     = 0x00000001 /* read only filesystem */
	MntSynchronous  = 0x00000002 /* filesystem written synchronously */
	MntNoExec       = 0x00000004 /* can't exec from filesystem */
	MntNoSuid       = 0x00000008 /* don't honor setuid bits on fs */
	MntUnion        = 0x00000020 /* union with underlying filesystem */
	MntAsync        = 0x00000040 /* filesystem written asynchronously */
	MntSuidDir      = 0x00100000 /* special handling of SUID on dirs */
	MntSoftDep      = 0x00200000 /* soft updates being done */
	MntNoSymFollow  = 0x00400000 /* do not follow symlinks */
	MntGEOMJournal  = 0x02000000 /* GEOM journal support enabled */
	MntMultilabel   = 0x04000000 /* MAC support for individual objects */
	MntACLs         = 0x08000000 /* ACL support enabled */
	MntNoATime      = 0x10000000 /* disable update of file access time */
	MntClusterRead  = 0x40000000 /* disable cluster read */
	MntClusterWrite = 0x80000000 /* disable cluster write */
	MntNFS4ACLs     = 0x00000010
)

func DiskPartitions(all bool) ([]DiskPartitionStat, error) {
	var ret []DiskPartitionStat

	count, err := unix.Getfsstat(nil, 1)
	if err != nil {
		return ret, err
	}
	fs := make([]unix.Statfs_t, count)
	_, err = unix.Getfsstat(fs, 1)
	for _, stat := range fs {
		opts := "rw"
		if stat.Flags&MntReadOnly != 0 {
			opts = "ro"
		}
		if stat.Flags&MntSynchronous != 0 {
			opts += ",sync"
		}
		if stat.Flags&MntNoExec != 0 {
			opts += ",noexec"
		}
		if stat.Flags&MntNoSuid != 0 {
			opts += ",nosuid"
		}
		if stat.Flags&MntUnion != 0 {
			opts += ",union"
		}
		if stat.Flags&MntAsync != 0 {
			opts += ",async"
		}
		if stat.Flags&MntSuidDir != 0 {
			opts += ",suiddir"
		}
		if stat.Flags&MntSoftDep != 0 {
			opts += ",softdep"
		}
		if stat.Flags&MntNoSymFollow != 0 {
			opts += ",nosymfollow"
		}
		if stat.Flags&MntGEOMJournal != 0 {
			opts += ",gjounalc"
		}
		if stat.Flags&MntMultilabel != 0 {
			opts += ",multilabel"
		}
		if stat.Flags&MntACLs != 0 {
			opts += ",acls"
		}
		if stat.Flags&MntNoATime != 0 {
			opts += ",noattime"
		}
		if stat.Flags&MntClusterRead != 0 {
			opts += ",nocluster"
		}
		if stat.Flags&MntClusterWrite != 0 {
			opts += ",noclusterw"
		}
		if stat.Flags&MntNFS4ACLs != 0 {
			opts += ",nfs4acls"
		}
		d := DiskPartitionStat{
			Device:     common.IntToString(stat.Mntfromname[:]),
			Mountpoint: common.IntToString(stat.Mntonname[:]),
			Fstype:     common.IntToString(stat.Fstypename[:]),
			Opts:       opts,
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	return nil, common.NotImplementedError
}
