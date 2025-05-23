package net

import (
	"context"
	"errors"
	"fmt"
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
type findDevsF func() ([]pcap.Interface, error)

func tracePackets(ctx context.Context, kind string) {
	devs := findActiveDevices(pcap.FindAllDevs)
	if len(devs) == 0 {
		errChan <- errors.New("no active network devices found")
		return
	}

	for _, dev := range devs {
		// one thread per interface
		go processDeviceMsgs(ctx, &dev, kind)
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

func processDeviceMsgs(ctx context.Context, dev *pcap.Interface, kind string) {
	var handle *pcap.Handle
	var err error

	if handle, err = pcap.OpenLive(dev.Name, 1600, false, 1*time.Second); err != nil {
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

		var errCnt uint64
		errLayer := p.ErrorLayer()
		if errLayer != nil {
			// What about errLayer.Error()?
			errCnt = 1
		}

		nBytes := uint64(len(p.Data()))
		var dstAddr, srcAddr Addr
		if decodeTCP(p, &srcAddr, &dstAddr) || decodeUDP(p, &srcAddr, &dstAddr) {
			if stat, ok := ProcConnMap[srcAddr]; ok {
				ProcConnMap[srcAddr].NetCounters.BytesSent += nBytes
				stat.NetCounters.PacketsSent++
				stat.NetCounters.Errout += errCnt
			} else if stat, ok := ProcConnMap[dstAddr]; ok {
				stat.NetCounters.BytesRecv += nBytes
				stat.NetCounters.PacketsRecv++
				stat.NetCounters.Errin += errCnt
			}
			// } else {
			// 	fmt.Printf("--- Not in the table: src=%v, dst=%v\n", srcAddr, dstAddr) //UC
			// }
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
