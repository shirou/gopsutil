// +build linux
// +build netlink

package net

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/shirou/gopsutil/internal/common"
	"github.com/shirou/gopsutil/net/netlink"
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
		return connectionsNetLink(tmap)
	}
	inodes, err = getProcInodes(root, pid, max)
	if len(inodes) == 0 {
		// no connection for the pid
		return []ConnectionStat{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cound not get pid(s), %d: %s", pid, err)
	}
	return statsFromInodes(root, pid, tmap, inodes, skipUids)
}

func connectionsNetLink(kinds []netConnectionKindType) ([]ConnectionStat, error) {
	var wait sync.WaitGroup
	reply := make([][]ConnectionStat, len(kinds))
	var retErrr error

	for i, kind := range kinds {
		if kind.family == syscall.AF_UNIX { // TODO: Unix Domain
			continue
		}
		wait.Add(1)
		go func(i int, k netConnectionKindType) {
			defer wait.Done()
			msgs, err := netlink.Connections(k.family, k.netlinkProto)
			if err != nil {
				retErrr = err
				return
			}
			t := make([]ConnectionStat, len(msgs))
			for i, msg := range msgs {
				conn := ConnectionStat{
					Family: uint32(msg.Family),
					Type:   uint32(k.sockType),
					Laddr:  Addr{msg.SrcIP().String(), uint32(msg.SrcPort())},
					Raddr:  Addr{msg.DstIP().String(), uint32(msg.DstPort())},
					Status: netlink.TCPState(msg.State).String(),
					Uids:   []int32{int32(msg.UID)},
					// Fd:     diag.Inode,
					// Pid:    c.pid,
				}
				t[i] = conn
			}
			reply[i] = t
		}(i, kind)
	}

	wait.Wait()

	ret := make([]ConnectionStat, 0)
	for i := range kinds {
		ret = append(ret, reply[i]...)
	}

	return ret, retErrr
}
