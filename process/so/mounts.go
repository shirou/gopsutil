// +build linux

package so

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// mount represents a line entry in /proc/<PID>/mountinfo
//
// From https://www.kernel.org/doc/html/latest/filesystems/proc.html#proc-pid-mountinfo-information-about-mounts
// 36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
// (1)(2)(3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

// (1) mount ID:  unique identifier of the mount (may be reused after umount)
// (2) parent ID:  ID of parent (or of self for the top of the mount tree)
// (3) major:minor:  value of st_dev for files on filesystem
// (4) root:  root of the mount within the filesystem
// (5) mount point:  mount point relative to the process's root
// (6) mount options:  per mount options
// (7) optional fields:  zero or more fields of the form "tag[:value]"
// (8) separator:  marks the end of the optional fields
// (9) filesystem type:  name of filesystem of the form "type[.subtype]"
// (10) mount source:  filesystem specific information or "none"
// (11) super options:  per super block options
type mount struct {
	dev        string
	root       string
	mountPoint string
}

type mountInfo struct {
	mounts []*mount
}

func getMountInfo(pidPath string, b *bufio.Reader) *mountInfo {
	f, err := os.Open(filepath.Join(pidPath, "mountinfo"))
	if err != nil {
		return nil
	}
	defer f.Close()
	b.Reset(f)

	return parseMountInfo(b)
}

func parseMountInfo(r *bufio.Reader) *mountInfo {
	mounts := make([]*mount, 0, 30)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}

		fields := bytes.Fields(line)
		if len(fields) < 10 {
			continue
		}

		mounts = append(mounts, &mount{
			dev:        string(fields[2]),
			root:       string(fields[3]),
			mountPoint: string(fields[4]),
		})
	}

	return &mountInfo{mounts: mounts}
}

// GetMount returns the mount where path resides, by searching the longest common mountPoint prefix
func (m *mountInfo) GetMount(path string) *mount {
	var match *mount
	for _, mount := range m.mounts {
		if strings.HasPrefix(path, mount.mountPoint) && (match == nil || len(mount.mountPoint) > len(match.mountPoint)) {
			match = mount
		}
	}
	return match
}
