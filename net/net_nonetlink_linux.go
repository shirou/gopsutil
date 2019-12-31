// +build linux
// +build !netlink

package net

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/internal/common"
)

func connectionsPidMaxWithoutUidsWithContext(ctx context.Context, kind string, pid int32, max int, skipUids bool) ([]ConnectionStat, error) {
	tmap, ok := netConnectionKindMap[kind]
	if !ok {
		return nil, fmt.Errorf("invalid kind, %s", kind)
	}
	root := common.HostProc()
	var err error
	var inodes map[string][]inodeMap
	if pid == 0 {
		inodes, err = getProcInodesAll(root, max)
	} else {
		inodes, err = getProcInodes(root, pid, max)
	}
	if len(inodes) == 0 {
		// no connection for the pid
		return []ConnectionStat{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cound not get pid(s), %d: %s", pid, err)
	}
	return statsFromInodes(root, pid, tmap, inodes, skipUids)
}
