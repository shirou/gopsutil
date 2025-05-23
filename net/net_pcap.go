package net

import (
	"context"
	"errors"
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

	CHECK_DONE_INTVL = 10 * time.Millisecond
)

func tracePackets(ctx context.Context, kind string) {
	devs := findActiveDevices(kind)
	if len(devs) == 0 {
		errChannel <- errors.New("no active network devices found")
		return
	}

	for _, dev := range devs {
		// one thread per interface
		go processDeviceMsgs(ctx, &dev, kind)
	}
}

func findActiveDevices(kind string) map[int]pcap.Interface {
	ret := make(map[int]pcap.Interface)

	devs, err := pcap.FindAllDevs()
	if err != nil {
		errChannel <- err
		return ret
	}

	for idx, dev := range devs {
		if (dev.Flags&PCAP_GTG_FLAGS) == PCAP_GTG_FLAGS && len(dev.Addresses) > 0 && isRightKind(dev, kind) { // UC exclude multicast ports?
			ret[idx+1] = dev
		}
	}

	return ret
}

func processDeviceMsgs(ctx context.Context, dev *pcap.Interface, kind string) {
	var handle *pcap.Handle
	var err error

	if handle, err = pcap.OpenLive(dev.Name, 1600, false, 1*time.Second); err != nil {
		errChannel <- err
		return
	}
	defer handle.Close()

	if handle.SetBPFFilter(kindToBPFFilter(kind)) != nil {
		errChannel <- err
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
			//UC err = errLayer.Error()
			errCnt = 1
		}

		nBytes := uint64(len(p.Data()))
		var dstAddr, srcAddr Addr
		if decodeTcp(p, &srcAddr, &dstAddr) || decodeUdp(p, &srcAddr, &dstAddr) {
			if stat, ok := ProcConnMap[srcAddr]; ok {
				stat.NetCounters.BytesSent += nBytes
				stat.NetCounters.PacketsSent++
				stat.NetCounters.Errout += errCnt
			} else if stat, ok := ProcConnMap[dstAddr]; ok {
				stat.NetCounters.BytesRecv += nBytes
				stat.NetCounters.PacketsRecv++
				stat.NetCounters.Errin += errCnt
			}
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

func isRightKind(_ pcap.Interface, _ string) bool {
	return true // UC WIP
}

func kindToBPFFilter(_ string) string {
	return "tcp||udp" // UC WIP
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

func decodeTcp(p gopacket.Packet, srcAddr *Addr, dstAddr *Addr) bool {
	var srcIp, dstIp string
	if !decodeIP(p, &srcIp, &dstIp) {
		return false
	}

	tcpLayer := p.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return false
	}

	tcp := tcpLayer.(*layers.TCP)
	*srcAddr = Addr{IP: srcIp, Port: uint32(tcp.SrcPort)}
	*dstAddr = Addr{IP: dstIp, Port: uint32(tcp.DstPort)}
	return true
}

func decodeUdp(p gopacket.Packet, srcAddr *Addr, dstAddr *Addr) bool {
	var srcIp, dstIp string
	if !decodeIP(p, &srcIp, &dstIp) {
		return false
	}

	udpLayer := p.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return false
	}

	udp := udpLayer.(*layers.UDP)
	*srcAddr = Addr{IP: srcIp, Port: uint32(udp.SrcPort)}
	*dstAddr = Addr{IP: dstIp, Port: uint32(udp.DstPort)}
	return true
}

func decodeIP(p gopacket.Packet, srcIp *string, dstIp *string) bool {
	ip4Layer := p.Layer(layers.LayerTypeIPv4)
	if ip4Layer != nil {
		ip := ip4Layer.(*layers.IPv4)
		*srcIp = ip.SrcIP.String()
		*dstIp = ip.DstIP.String()
		return true
	}

	ip6Layer := p.Layer(layers.LayerTypeIPv6)
	if ip6Layer != nil {
		ip := ip6Layer.(*layers.IPv6)
		*srcIp = ip.SrcIP.String()
		*dstIp = ip.DstIP.String()
		return true
	}

	return false
}
