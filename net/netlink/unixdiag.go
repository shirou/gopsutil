// +build linux

package netlink

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

// Show flags in unix diag request.
// https://github.com/torvalds/linux/blob/v4.0/include/uapi/linux/unix_diag.h#L16
const (
	UDIAG_SHOW_NAME    = 0x1
	UDIAG_SHOW_VFS     = 0x00000002 /* show VFS inode info */
	UDIAG_SHOW_PEER    = 0x00000004 /* show peer socket info */
	UDIAG_SHOW_ICONS   = 0x00000008 /* show pending connections */
	UDIAG_SHOW_RQLEN   = 0x00000010 /* show skb receive queue len */
	UDIAG_SHOW_MEMINFO = 0x00000020 /* show memory info of a socket */
)

const (
	/* UNIX_DIAG_NONE, standard nl API requires this attribute!  */
	UNIX_DIAG_NAME = iota
	UNIX_DIAG_VFS
	UNIX_DIAG_PEER
	UNIX_DIAG_ICONS
	UNIX_DIAG_RQLEN
	UNIX_DIAG_MEMINFO
	UNIX_DIAG_SHUTDOWN
)

// UnixDiag sends the given netlink request
func UnixDiag(request syscall.NetlinkMessage) ([]*UnixDiagMsgExtended, error) {
	return UnixDiagWithBuf(request, nil, nil)
}

// UnixDiagWithBuf sends the given netlink request parses the responsesor
// debugging).
func UnixDiagWithBuf(request syscall.NetlinkMessage, readBuf []byte, resp io.Writer) ([]*UnixDiagMsgExtended, error) {
	s, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_INET_DIAG) // same as NETLINK_SOCK_DIAG
	if err != nil {
		return nil, err
	}
	defer syscall.Close(s)

	lsa := &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK}
	if err := syscall.Sendto(s, serialize(request), 0, lsa); err != nil {
		return nil, err
	}

	if len(readBuf) == 0 {
		// Default size used in libnl.
		readBuf = make([]byte, os.Getpagesize())
	}

	var unixDiagMsgsExtended []*UnixDiagMsgExtended
done:
	for {
		buf := readBuf
		nr, _, err := syscall.Recvfrom(s, buf, 0)
		if err != nil {
			return nil, err
		}
		if nr < syscall.NLMSG_HDRLEN {
			return nil, syscall.EINVAL
		}

		buf = buf[:nr]

		// Dump raw data for inspection purposes.
		if resp != nil {
			if _, err := resp.Write(buf); err != nil {
				return nil, err
			}
		}

		msgs, err := syscall.ParseNetlinkMessage(buf)
		if err != nil {
			return nil, err
		}

		dupCheckMap := make(map[string]struct{})
		for _, m := range msgs {
			if m.Header.Type == syscall.NLMSG_DONE {
				break done
			}
			if m.Header.Type == syscall.NLMSG_ERROR {
				return nil, ParseNetlinkError(m.Data)
			}
			if m.Header.Type != SOCK_DIAG_BY_FAMILY {
				return nil, fmt.Errorf("unexpected nlmsg_type %d", m.Header.Type)
			}

			extended, err := ParseUnixDiagMsg(m.Data, int(m.Header.Len))
			if err != nil {
				return nil, err
			}
			if extended.Path == "" {
				continue
			}
			if _, exists := dupCheckMap[extended.Path]; exists {
				continue
			}
			dupCheckMap[extended.Path] = struct{}{}

			unixDiagMsgsExtended = append(unixDiagMsgsExtended, extended)
		}
	}
	return unixDiagMsgsExtended, nil
}

var sizeofUnixDiagReq = int(unsafe.Sizeof(UnixDiagReq{}))

// UnixDiagReq is used to request diagnostic data.
// https://github.com/torvalds/linux/blob/v4.0/include/uapi/linux/unix_diag.h#L6
type UnixDiagReq struct {
	Family   uint8
	Protocol uint8
	Pad      uint16
	States   uint32
	Ino      uint32 // inode number
	Show     uint32 // report information status
	Cookie   [2]uint32
}

func (r UnixDiagReq) toWireFormat() []byte {
	buf := bytes.NewBuffer(make([]byte, sizeofUnixDiagReq))
	buf.Reset()
	if err := binary.Write(buf, byteOrder, r); err != nil {
		// This never returns an error.
		panic(err)
	}
	return buf.Bytes()
}

// NewUnixDiagReq returns a new NetlinkMessage whose payload is an
// UnixDiagReq. Callers should set their own sequence number in the returned
// message header.
func NewUnixDiagReq() syscall.NetlinkMessage {
	hdr := syscall.NlMsghdr{
		Type:  uint16(SOCK_DIAG_BY_FAMILY),
		Flags: uint16(syscall.NLM_F_DUMP | syscall.NLM_F_REQUEST),
		Pid:   uint32(0),
	}
	req := UnixDiagReq{
		Family: syscall.AF_UNIX,
		States: AllTCPStates,
		//		Show:   uint32(UDIAG_SHOW_NAME | UDIAG_SHOW_PEER),
		Show: uint32(UDIAG_SHOW_NAME),
	}

	return syscall.NetlinkMessage{Header: hdr, Data: req.toWireFormat()}
}

// Response messages.

// UnixDiagMsg (unix_diag_msg) represents return message.
// https://github.com/torvalds/linux/blob/v4.0/include/uapi/linux/unix_diag.h#L23
type UnixDiagMsg struct {
	Family uint8
	Type   uint8
	State  uint8
	Pad    uint8

	Inode  uint32
	Cookie [2]uint32
}

// UnixDiagMsgExtended is a extended struct of UnixDiagMsg. Because UnixDiagMsg is used to parse
// so it can not include other information.
type UnixDiagMsgExtended struct {
	*UnixDiagMsg
	Path string
}

var sizeofUnixDiagMsg = int(unsafe.Sizeof(UnixDiagMsg{}))

// ParseUnixDiagMsg parse an UnixDiagMsg from a byte slice. It assumes the
// UnixDiagMsg starts at the beginning of b. Invoke this method to parse the
// payload of a netlink response.
func ParseUnixDiagMsg(b []byte, dataLen int) (*UnixDiagMsgExtended, error) {
	r := bytes.NewReader(b)
	unixDiagMsg := &UnixDiagMsg{}
	if err := binary.Read(r, byteOrder, unixDiagMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal unix_diag_msg")
	}

	// parse rest of buffer
	attrs, err := ParseNetlinkRouteAttr(b[sizeofUnixDiagMsg:])
	if err != nil {
		return nil, err
	}
	extended := &UnixDiagMsgExtended{unixDiagMsg, ""}
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case UNIX_DIAG_NAME:
			if attr.Value[0] == 0 {
				extended.Path = "@" + string(attr.Value[1:])
			} else {
				extended.Path = string(attr.Value)
			}
		}
	}

	return extended, nil
}

// ParseNetlinkRouteAttr parse Route attr from NetlinkMessage.Data
func ParseNetlinkRouteAttr(b []byte) ([]syscall.NetlinkRouteAttr, error) {
	var attrs []syscall.NetlinkRouteAttr
	for len(b) >= syscall.SizeofRtAttr {
		a, vbuf, err := netlinkRouteAttrAndValue(b)
		if err != nil {
			return nil, err
		}
		ra := syscall.NetlinkRouteAttr{Attr: *a, Value: bytes.Trim(vbuf[:RtaPayload(a)], "\x00")}
		attrs = append(attrs, ra)
		b = b[rtaAlignOf(int(ra.Attr.Len)):]
	}
	return attrs, nil
}

// modifed from syscall/netlink_linux.go
func netlinkRouteAttrAndValue(b []byte) (*syscall.RtAttr, []byte, error) {
	a := (*syscall.RtAttr)(unsafe.Pointer(&b[0]))
	if int(a.Len) < syscall.SizeofRtAttr || int(a.Len) > len(b) {
		return nil, nil, syscall.EINVAL
	}
	// fmt.Printf("netlinkRouteAttrAndValue: type=%d, len=%d, %d\n", a.Type, a.Len, RtaPayload(a))
	return a, b[syscall.SizeofRtAttr:a.Len], nil
}

// Round the length of a netlink route attribute up to align it
// properly.
func rtaAlignOf(attrlen int) int {
	return (attrlen + syscall.RTA_ALIGNTO - 1) & ^(syscall.RTA_ALIGNTO - 1)
}

// RtaLength is a implementation of RTA_LENGTH macro.
func RtaLength(len int) int {
	return rtaAlignOf(syscall.SizeofRtAttr) + len
}

// RtaPayload is a implementation of RTA_PAYLOAD macro.
func RtaPayload(rta *syscall.RtAttr) int {
	return int(rta.Len) - RtaLength(0)
}
