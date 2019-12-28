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

func getInetConnections(wait *sync.WaitGroup, i int, k netConnectionKindType, reply [][]ConnectionStat) {
	defer wait.Done()
	msgs, err := netlink.InetConnections(k.family, k.netlinkProto)
	if err != nil {
		// just ignored
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
			// Pid:    msg.Pid,
			// Fd:     diag.Inode,
		}
		t[i] = conn
	}
	reply[i] = t
}

func getUnixConnections(wait *sync.WaitGroup, i int, reply [][]ConnectionStat) {
	defer wait.Done()
	msgs, err := netlink.UnixConnections()
	if err != nil {
		// just ignored
		return
	}
	t := make([]ConnectionStat, len(msgs))
	for i, msg := range msgs {
		conn := ConnectionStat{
			Family: uint32(msg.Family),
			Status: netlink.TCPState(msg.State).String(),
			Uids:   []int32{},
			Laddr:  Addr{
				IP: msg.Path,
			},
		}
		t[i] = conn
	}
	reply[i] = t
}

func connectionsNetLink(kinds []netConnectionKindType) ([]ConnectionStat, error) {
	wait := &sync.WaitGroup{}
	reply := make([][]ConnectionStat, len(kinds))
	var retErrr error

	for i, kind := range kinds {
		wait.Add(1)
		if kind.family == syscall.AF_UNIX {
			go getUnixConnections(wait, i, reply)
		} else {
			go getInetConnections(wait, i, kind, reply)
		}
	}

	wait.Wait()

	ret := make([]ConnectionStat, 0)
	for i := range kinds {
		ret = append(ret, reply[i]...)
	}

	return ret, retErrr
}
