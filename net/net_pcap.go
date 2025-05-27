package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	PCAP_IF_LOOPBACK                         = uint32(0x00000001)
	PCAP_IF_UP                               = uint32(0x00000002)
	PCAP_IF_RUNNING                          = uint32(0x00000004)
	PCAP_IF_WIRELESS                         = uint32(0x00000008)
	PCAP_IF_CONNECTION_STATUS_UNKNOWN        = uint32(0x00000000)
	PCAP_IF_CONNECTION_STATUS_CONNECTED      = uint32(0x00000010)
	PCAP_IF_CONNECTION_STATUS_DISCONNECTED   = uint32(0x00000020)
	PCAP_IF_CONNECTION_STATUS_NOT_APPLICABLE = uint32(0x00000030)
	PCAP_GTG_FLAGS                           = uint32(PCAP_IF_UP | PCAP_IF_RUNNING | PCAP_IF_CONNECTION_STATUS_CONNECTED)
)

const (
	CHECK_DONE_INTVL = 10 * time.Millisecond
)

// For unit test mocking
type (
	findDevsF func() ([]pcap.Interface, error)
	openLiveF func(device string, snaplen int32, promisc bool, timeout time.Duration) (*pcap.Handle, error)
)

func tracePackets(ctx context.Context, kind string) {
	devs := findActiveDevices(pcap.FindAllDevs)
	if len(devs) == 0 {
		errChan <- errors.New("no active network devices found")
		return
	}

	for _, dev := range devs {
		// one thread per interface
		go processDeviceMsgs(ctx, &dev, kind, pcap.OpenLive)
	}
}

func findActiveDevices(findDevs findDevsF) []pcap.Interface {
	ret := make([]pcap.Interface, 0)

	devs, err := findDevs()
	if err != nil {
		errChan <- err
		return ret
	}

	for _, dev := range devs {
		if (dev.Flags&PCAP_GTG_FLAGS) == PCAP_GTG_FLAGS && len(dev.Addresses) > 0 {
			ret = append(ret, dev)
		}
	}

	return ret
}

func processDeviceMsgs(ctx context.Context, dev *pcap.Interface, kind string, openLive openLiveF) {
	var handle *pcap.Handle
	var err error

	if handle, err = openLive(dev.Name, 1600, false, 1*time.Second); err != nil {
		errChan <- err
		return
	}
	defer handle.Close()

	if handle.SetBPFFilter(kindToBPFFilter(kind)) != nil {
		errChan <- err
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetCh := packetSource.Packets()

	for keepTracing(ctx) {
		p := waitNextPacket(packetCh)
		if p == nil {
			continue
		}

		processPacket(p, dev)
	}
}

func addTransientConn(srcAddr *Addr, dstAddr *Addr, dev *pcap.Interface) *Addr {
	watchLock.Lock()
	defer watchLock.Unlock()

	lclAddr, rmtAddr, err := sortAddresses(srcAddr, dstAddr, dev)
	if err != nil {
		errChan <- err
		return nil
	}

	ProcConnMap[*lclAddr] = &ProcNetStat{Pid: -1, NetCounters: IOCountersStat{}, RemoteAddr: *rmtAddr, LastUpdate: time.Now()}
	return lclAddr
}

func updateIfName(stat *ProcNetStat, iName string) {
	if len(stat.NetCounters.Name) == 0 {
		stat.NetCounters.Name = iName
	}
}

func ipMatch(ip net.IP, nicAddr pcap.InterfaceAddress) bool {
	return slices.Equal(ip.Mask(nicAddr.Netmask), nicAddr.IP.Mask(nicAddr.Netmask))
}

// Figures which address is local and which remote
func sortAddresses(addr1 *Addr, addr2 *Addr, dev *pcap.Interface) (*Addr, *Addr, error) {
	ip1 := net.ParseIP(addr1.IP)
	ip2 := net.ParseIP(addr2.IP)
	for _, nicAddr := range dev.Addresses {
		switch {
		case ipMatch(ip1, nicAddr) && !ipMatch(ip2, nicAddr):
			return addr1, addr2, nil
		case ipMatch(ip2, nicAddr) && !ipMatch(ip1, nicAddr):
			return addr2, addr1, nil
		case ipMatch(ip1, nicAddr) && addr1.Port >= addr2.Port:
			return addr1, addr2, nil
		case ipMatch(ip2, nicAddr) && addr1.Port < addr2.Port:
			return addr2, addr1, nil
		}
	}
	return nil, nil, fmt.Errorf("addresses %v, %v don't belong to the interface %s", addr1, addr2, dev.Name)
}

func processPacket(p gopacket.Packet, dev *pcap.Interface) {
	var errCnt uint64
	errLayer := p.ErrorLayer()
	if errLayer != nil {
		// What about errLayer.Error()?
		errCnt = 1
	}

	nBytes := uint64(len(p.Data()))
	var dstAddr, srcAddr Addr
	if decodeTCP(p, &srcAddr, &dstAddr) || decodeUDP(p, &srcAddr, &dstAddr) {
		statOut, isOut := ProcConnMap[srcAddr]
		statIn, isIn := ProcConnMap[dstAddr]
		if !isIn && !isOut {
			addTransientConn(&srcAddr, &dstAddr, dev)
			statOut, isOut = ProcConnMap[srcAddr]
			statIn, isIn = ProcConnMap[dstAddr]
		}

		if isOut {
			statOut.NetCounters.BytesSent += nBytes
			statOut.NetCounters.PacketsSent++
			statOut.NetCounters.Errout += errCnt
			statOut.LastUpdate = time.Now()
			updateIfName(statOut, dev.Name)
		} else if isIn {
			statIn.NetCounters.BytesRecv += nBytes
			statIn.NetCounters.PacketsRecv++
			statIn.NetCounters.Errin += errCnt
			statIn.LastUpdate = time.Now()
			updateIfName(statIn, dev.Name)
		}
	}
}

func keepTracing(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func kindToBPFFilter(kind string) string {
	switch kind {
	case "all", "inet":
		return "tcp || udp"
	case "tcp":
		return "tcp"
	case "tcp4":
		return "ip && tcp"
	case "tcp6":
		return "ip6 && tcp"
	case "udp":
		return "udp"
	case "udp4":
		return "ip && udp"
	case "udp6":
		return "ip6 && udp"
	case "inet4":
		return "ip"
	case "inet6":
		return "ip6"
	}
	errChan <- fmt.Errorf("unknown network kind '%s', capturing all", kind)
	return "tcp || udp"
}

func waitNextPacket(packetCh chan gopacket.Packet) gopacket.Packet {
	select {
	case p := <-packetCh:
		return p
	default:
		time.Sleep(CHECK_DONE_INTVL)
		return nil
	}
}

func decodeTCP(p gopacket.Packet, srcAddr *Addr, dstAddr *Addr) bool {
	var srcIP, dstIP string
	if !decodeIP(p, &srcIP, &dstIP) {
		return false
	}

	tcpLayer := p.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return false
	}

	tcp := tcpLayer.(*layers.TCP)
	*srcAddr = Addr{IP: srcIP, Port: uint32(tcp.SrcPort)}
	*dstAddr = Addr{IP: dstIP, Port: uint32(tcp.DstPort)}
	return true
}

func decodeUDP(p gopacket.Packet, srcAddr *Addr, dstAddr *Addr) bool {
	var srcIP, dstIP string
	if !decodeIP(p, &srcIP, &dstIP) {
		return false
	}

	udpLayer := p.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return false
	}

	udp := udpLayer.(*layers.UDP)
	*srcAddr = Addr{IP: srcIP, Port: uint32(udp.SrcPort)}
	*dstAddr = Addr{IP: dstIP, Port: uint32(udp.DstPort)}
	return true
}

func decodeIP(p gopacket.Packet, srcIP *string, dstIP *string) bool {
	ip4Layer := p.Layer(layers.LayerTypeIPv4)
	if ip4Layer != nil {
		ip := ip4Layer.(*layers.IPv4)
		*srcIP = ip.SrcIP.String()
		*dstIP = ip.DstIP.String()
		return true
	}

	ip6Layer := p.Layer(layers.LayerTypeIPv6)
	if ip6Layer != nil {
		ip := ip6Layer.(*layers.IPv6)
		*srcIP = ip.SrcIP.String()
		*dstIP = ip.DstIP.String()
		return true
	}

	return false
}
