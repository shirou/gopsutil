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
	ADFS_SUPER_MAGIC      = 0xadf5
	AFFS_SUPER_MAGIC      = 0xADFF
	BEFS_SUPER_MAGIC      = 0x42465331
	BFS_MAGIC             = 0x1BADFACE
	CIFS_MAGIC_NUMBER     = 0xFF534D42
	CODA_SUPER_MAGIC      = 0x73757245
	COH_SUPER_MAGIC       = 0x012FF7B7
	CRAMFS_MAGIC          = 0x28cd3d45
	DEVFS_SUPER_MAGIC     = 0x1373
	EFS_SUPER_MAGIC       = 0x00414A53
	EXT_SUPER_MAGIC       = 0x137D
	EXT2_OLD_SUPER_MAGIC  = 0xEF51
	EXT2_SUPER_MAGIC      = 0xEF53
	EXT3_SUPER_MAGIC      = 0xEF53
	EXT4_SUPER_MAGIC      = 0xEF53
	HFS_SUPER_MAGIC       = 0x4244
	HPFS_SUPER_MAGIC      = 0xF995E849
	HUGETLBFS_MAGIC       = 0x958458f6
	ISOFS_SUPER_MAGIC     = 0x9660
	JFFS2_SUPER_MAGIC     = 0x72b6
	JFS_SUPER_MAGIC       = 0x3153464a
	MINIX_SUPER_MAGIC     = 0x137F /* orig. minix */
	MINIX_SUPER_MAGIC2    = 0x138F /* 30 char minix */
	MINIX2_SUPER_MAGIC    = 0x2468 /* minix V2 */
	MINIX2_SUPER_MAGIC2   = 0x2478 /* minix V2, 30 char names */
	MSDOS_SUPER_MAGIC     = 0x4d44
	NCP_SUPER_MAGIC       = 0x564c
	NFS_SUPER_MAGIC       = 0x6969
	NTFS_SB_MAGIC         = 0x5346544e
	OPENPROM_SUPER_MAGIC  = 0x9fa1
	PROC_SUPER_MAGIC      = 0x9fa0
	QNX4_SUPER_MAGIC      = 0x002f
	REISERFS_SUPER_MAGIC  = 0x52654973
	ROMFS_MAGIC           = 0x7275
	SMB_SUPER_MAGIC       = 0x517B
	SYSV2_SUPER_MAGIC     = 0x012FF7B6
	SYSV4_SUPER_MAGIC     = 0x012FF7B5
	TMPFS_MAGIC           = 0x01021994
	UDF_SUPER_MAGIC       = 0x15013346
	UFS_MAGIC             = 0x00011954
	USBDEVICE_SUPER_MAGIC = 0x9fa2
	VXFS_SUPER_MAGIC      = 0xa501FCF5
	XENIX_SUPER_MAGIC     = 0x012FF7B4
	XFS_SUPER_MAGIC       = 0x58465342
	_XIAFS_SUPER_MAGIC    = 0x012FD16D
)

var fsTypeMap = map[int64]string{
	AFFS_SUPER_MAGIC:      "affs",
	COH_SUPER_MAGIC:       "coh",
	DEVFS_SUPER_MAGIC:     "devfs",
	EXT2_OLD_SUPER_MAGIC:  "old ext2",
	EXT2_SUPER_MAGIC:      "ext2",
	EXT3_SUPER_MAGIC:      "ext3",
	EXT4_SUPER_MAGIC:      "ext4",
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
	SMB_SUPER_MAGIC:       "smb",
	SYSV2_SUPER_MAGIC:     "sysv2",
	SYSV4_SUPER_MAGIC:     "sysv4",
	UFS_MAGIC:             "ufs",
	USBDEVICE_SUPER_MAGIC: "usb",
	VXFS_SUPER_MAGIC:      "vxfs",
	XENIX_SUPER_MAGIC:     "xenix",
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
