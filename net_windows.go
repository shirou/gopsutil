// +build windows

package gopsutil

import (
	"errors"
	"net"
	"os"
	"syscall"
	"unsafe"
)

var (
	modiphlpapi             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")
)

const (
	TCP_TABLE_BASIC_LISTENER = iota
	TCP_TABLE_BASIC_CONNECTIONS
	TCP_TABLE_BASIC_ALL
	TCP_TABLE_OWNER_PID_LISTENER
	TCP_TABLE_OWNER_PID_CONNECTIONS
	TCP_TABLE_OWNER_PID_ALL
	TCP_TABLE_OWNER_MODULE_LISTENER
	TCP_TABLE_OWNER_MODULE_CONNECTIONS
	TCP_TABLE_OWNER_MODULE_ALL
)

func NetIOCounters(pernic bool) ([]NetIOCountersStat, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ai, err := getAdapterList()
	if err != nil {
		return nil, err
	}
	var ret []NetIOCountersStat

	for _, ifi := range ifs {
		name := ifi.Name
		for ; ai != nil; ai = ai.Next {
			name = bytePtrToString(&ai.Description[0])
			c := NetIOCountersStat{
				Name: name,
			}

			row := syscall.MibIfRow{Index: ai.Index}
			e := syscall.GetIfEntry(&row)
			if e != nil {
				return nil, os.NewSyscallError("GetIfEntry", e)
			}
			c.BytesSent = uint64(row.OutOctets)
			c.BytesRecv = uint64(row.InOctets)
			c.PacketsSent = uint64(row.OutUcastPkts)
			c.PacketsRecv = uint64(row.InUcastPkts)
			c.Errin = uint64(row.InErrors)
			c.Errout = uint64(row.OutErrors)
			c.Dropin = uint64(row.InDiscards)
			c.Dropout = uint64(row.OutDiscards)

			ret = append(ret, c)
		}
	}
	return ret, nil
}

// Return a list of network connections opened by a process
func NetConnections(kind string) ([]NetConnectionStat, error) {
	var ret []NetConnectionStat

	return ret, errors.New("not implemented yet")
}

// borrowed from src/pkg/net/interface_windows.go
func getAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}
