// +build linux

package disk

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	common "github.com/shirou/gopsutil/common"
)

const (
	SectorSize = 512
)
const (
	// magic.h
	ADFS_SUPER_MAGIC     = 0xadf5
	AFFS_SUPER_MAGIC     = 0xadff
	AFS_SUPER_MAGIC      = 0x5346414F
	AUTOFS_SUPER_MAGIC   = 0x0187
	CODA_SUPER_MAGIC     = 0x73757245
	CRAMFS_MAGIC         = 0x28cd3d45 /* some random number */
	CRAMFS_MAGIC_WEND    = 0x453dcd28 /* magic number with the wrong endianess */
	DEBUGFS_MAGIC        = 0x64626720
	SECURITYFS_MAGIC     = 0x73636673
	SELINUX_MAGIC        = 0xf97cff8c
	SMACK_MAGIC          = 0x43415d53 /* "SMAC" */
	RAMFS_MAGIC          = 0x858458f6 /* some random number */
	TMPFS_MAGIC          = 0x01021994
	HUGETLBFS_MAGIC      = 0x958458f6 /* some random number */
	SQUASHFS_MAGIC       = 0x73717368
	ECRYPTFS_SUPER_MAGIC = 0xf15f
	EFS_SUPER_MAGIC      = 0x414A53
	EXT2_SUPER_MAGIC     = 0xEF53
	EXT3_SUPER_MAGIC     = 0xEF53
	XENFS_SUPER_MAGIC    = 0xabba1974
	EXT4_SUPER_MAGIC     = 0xEF53
	BTRFS_SUPER_MAGIC    = 0x9123683E
	NILFS_SUPER_MAGIC    = 0x3434
	F2FS_SUPER_MAGIC     = 0xF2F52010
	HPFS_SUPER_MAGIC     = 0xf995e849
	ISOFS_SUPER_MAGIC    = 0x9660
	JFFS2_SUPER_MAGIC    = 0x72b6
	PSTOREFS_MAGIC       = 0x6165676C
	EFIVARFS_MAGIC       = 0xde5e81e4
	HOSTFS_SUPER_MAGIC   = 0x00c0ffee

	MINIX_SUPER_MAGIC   = 0x137F /* minix v1 fs, 14 char names */
	MINIX_SUPER_MAGIC2  = 0x138F /* minix v1 fs, 30 char names */
	MINIX2_SUPER_MAGIC  = 0x2468 /* minix v2 fs, 14 char names */
	MINIX2_SUPER_MAGIC2 = 0x2478 /* minix v2 fs, 30 char names */
	MINIX3_SUPER_MAGIC  = 0x4d5a /* minix v3 fs, 60 char names */

	MSDOS_SUPER_MAGIC    = 0x4d44 /* MD */
	NCP_SUPER_MAGIC      = 0x564c /* Guess, what = 0x564c is :-) */
	NFS_SUPER_MAGIC      = 0x6969
	OPENPROM_SUPER_MAGIC = 0x9fa1
	QNX4_SUPER_MAGIC     = 0x002f     /* qnx4 fs detection */
	QNX6_SUPER_MAGIC     = 0x68191122 /* qnx6 fs detection */

	REISERFS_SUPER_MAGIC = 0x52654973 /* used by gcc */
	/* used by file system utilities that
	   look at the superblock, etc.  */
	// REISERFS_SUPER_MAGIC_STRING	"ReIsErFs"
	// REISER2FS_SUPER_MAGIC_STRING	"ReIsEr2Fs"
	// REISER2FS_JR_SUPER_MAGIC_STRING	"ReIsEr3Fs"
	SMB_SUPER_MAGIC       = 0x517B
	CGROUP_SUPER_MAGIC    = 0x27e0eb
	STACK_END_MAGIC       = 0x57AC6E9D
	TRACEFS_MAGIC         = 0x74726163
	V9FS_MAGIC            = 0x01021997
	BDEVFS_MAGIC          = 0x62646576
	BINFMTFS_MAGIC        = 0x42494e4d
	DEVPTS_SUPER_MAGIC    = 0x1cd1
	FUTEXFS_SUPER_MAGIC   = 0xBAD1DEA
	PIPEFS_MAGIC          = 0x50495045
	PROC_SUPER_MAGIC      = 0x9fa0
	SOCKFS_MAGIC          = 0x534F434B
	SYSFS_MAGIC           = 0x62656572
	USBDEVICE_SUPER_MAGIC = 0x9fa2
	MTD_INODE_FS_MAGIC    = 0x11307854
	ANON_INODE_FS_MAGIC   = 0x09041934
	BTRFS_TEST_MAGIC      = 0x73727279
	NSFS_MAGIC            = 0x6e736673
)

var fsTypeMap = map[int64]string{
	AFFS_SUPER_MAGIC:     "affs",
	BTRFS_SUPER_MAGIC:    "btrfs",
	COH_SUPER_MAGIC:      "coh",
	DEVFS_SUPER_MAGIC:    "devfs",
	EXT2_OLD_SUPER_MAGIC: "old ext2",
	EXT2_SUPER_MAGIC:     "ext2",
	//EXT3_SUPER_MAGIC:      "ext3", // TODO: how to identify?
	//EXT4_SUPER_MAGIC:      "ext4",
	TMPFS_MAGIC:           "tmpfs",
	HFS_SUPER_MAGIC:       "hfs",
	HPFS_SUPER_MAGIC:      "hpfs",
	ISOFS_SUPER_MAGIC:     "isofs",
	MINIX2_SUPER_MAGIC:    "minix v2",
	MINIX2_SUPER_MAGIC2:   "minix v2 30 char",
	MINIX_SUPER_MAGIC:     "minix",
	MINIX_SUPER_MAGIC2:    "minix 30 char",
	MSDOS_SUPER_MAGIC:     "msdos",
	NCP_SUPER_MAGIC:       "ncp",
	NFS_SUPER_MAGIC:       "nfs",
	NTFS_SB_MAGIC:         "ntfs",
	PROC_SUPER_MAGIC:      "proc",
	REISERFS_SUPER_MAGIC:  "reiserfs",
	SMB_SUPER_MAGIC:       "smb",
	SYSV2_SUPER_MAGIC:     "sysv2",
	SYSV4_SUPER_MAGIC:     "sysv4",
	UFS_MAGIC:             "ufs",
	USBDEVICE_SUPER_MAGIC: "usb",
	VXFS_SUPER_MAGIC:      "vxfs",
	XENIX_SUPER_MAGIC:     "xenix",
	XENFS_SUPER_MAGIC:     "xenfs",
	XFS_SUPER_MAGIC:       "xfs",
	_XIAFS_SUPER_MAGIC:    "xiafs",
}

// Get disk partitions.
// should use setmntent(3) but this implement use /etc/mtab file
func DiskPartitions(all bool) ([]DiskPartitionStat, error) {

	filename := "/etc/mtab"
	lines, err := common.ReadLines(filename)
	if err != nil {
		return nil, err
	}

	ret := make([]DiskPartitionStat, 0, len(lines))

	for _, line := range lines {
		fields := strings.Fields(line)
		d := DiskPartitionStat{
			Device:     fields[0],
			Mountpoint: fields[1],
			Fstype:     fields[2],
			Opts:       fields[3],
		}
		ret = append(ret, d)
	}

	return ret, nil
}

func DiskIOCounters() (map[string]DiskIOCountersStat, error) {
	filename := "/proc/diskstats"
	lines, err := common.ReadLines(filename)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]DiskIOCountersStat, 0)
	empty := DiskIOCountersStat{}

	for _, line := range lines {
		fields := strings.Fields(line)
		name := fields[2]
		reads, err := strconv.ParseUint((fields[3]), 10, 64)
		if err != nil {
			return ret, err
		}
		rbytes, err := strconv.ParseUint((fields[5]), 10, 64)
		if err != nil {
			return ret, err
		}
		rtime, err := strconv.ParseUint((fields[6]), 10, 64)
		if err != nil {
			return ret, err
		}
		writes, err := strconv.ParseUint((fields[7]), 10, 64)
		if err != nil {
			return ret, err
		}
		wbytes, err := strconv.ParseUint((fields[9]), 10, 64)
		if err != nil {
			return ret, err
		}
		wtime, err := strconv.ParseUint((fields[10]), 10, 64)
		if err != nil {
			return ret, err
		}
		iotime, err := strconv.ParseUint((fields[12]), 10, 64)
		if err != nil {
			return ret, err
		}
		d := DiskIOCountersStat{
			ReadBytes:  rbytes * SectorSize,
			WriteBytes: wbytes * SectorSize,
			ReadCount:  reads,
			WriteCount: writes,
			ReadTime:   rtime,
			WriteTime:  wtime,
			IoTime:     iotime,
		}
		if d == empty {
			continue
		}
		d.Name = name

		d.SerialNumber = GetDiskSerialNumber(name)
		ret[name] = d
	}
	return ret, nil
}

func GetDiskSerialNumber(name string) string {
	n := fmt.Sprintf("--name=%s", name)
	out, err := exec.Command("/sbin/udevadm", "info", "--query=property", n).Output()

	// does not return error, just an empty string
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		values := strings.Split(line, "=")
		if len(values) < 2 || values[0] != "ID_SERIAL" {
			// only get ID_SERIAL, not ID_SERIAL_SHORT
			continue
		}
		return values[1]
	}
	return ""
}

func getFsType(stat syscall.Statfs_t) string {
	t := stat.Type
	ret, ok := fsTypeMap[t]
	if !ok {
		return ""
	}
	return ret
}
