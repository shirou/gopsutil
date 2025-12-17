// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package net

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v4/internal/common"
)

// Deprecated: use process.PidsWithContext instead
func PidsWithContext(_ context.Context) ([]int32, error) {
	return nil, common.ErrNotImplementedError
}

func IOCountersByFileWithContext(ctx context.Context, pernic bool, _ string) ([]IOCountersStat, error) {
	return IOCountersWithContext(ctx, pernic)
}

func FilterCountersWithContext(_ context.Context) ([]FilterStat, error) {
	return nil, common.ErrNotImplementedError
}

func ConntrackStatsWithContext(_ context.Context, _ bool) ([]ConntrackStat, error) {
	return nil, common.ErrNotImplementedError
}

func ProtoCountersWithContext(_ context.Context, _ []string) ([]ProtoCountersStat, error) {
	return nil, common.ErrNotImplementedError
}

func parseNetstatNetLine(line string) (ConnectionStat, error) {
	f := strings.Fields(line)
	if len(f) < 5 {
		return ConnectionStat{}, fmt.Errorf("wrong line,%s", line)
	}

	var netType, netFamily uint32
	switch f[0] {
	case "tcp", "tcp4":
		netType = syscall.SOCK_STREAM
		netFamily = syscall.AF_INET
	case "udp", "udp4":
		netType = syscall.SOCK_DGRAM
		netFamily = syscall.AF_INET
	case "tcp6":
		netType = syscall.SOCK_STREAM
		netFamily = syscall.AF_INET6
	case "udp6":
		netType = syscall.SOCK_DGRAM
		netFamily = syscall.AF_INET6
	default:
		return ConnectionStat{}, fmt.Errorf("unknown type, %s", f[0])
	}

	laddr, raddr, err := parseNetstatAddr(f[3], f[4], netFamily)
	if err != nil {
		return ConnectionStat{}, fmt.Errorf("failed to parse netaddr, %s %s", f[3], f[4])
	}

	n := ConnectionStat{
		Fd:     uint32(0), // not supported
		Family: uint32(netFamily),
		Type:   uint32(netType),
		Laddr:  laddr,
		Raddr:  raddr,
		Pid:    int32(0), // not supported
	}
	if len(f) == 6 {
		n.Status = f[5]
	}

	return n, nil
}

var portMatch = regexp.MustCompile(`(.*)\.(\d+)$`)

func parseAddr(l string, family uint32) (Addr, error) {
	matches := portMatch.FindStringSubmatch(l)
	if matches == nil {
		return Addr{}, fmt.Errorf("wrong addr, %s", l)
	}
	host := matches[1]
	port := matches[2]
	if host == "*" {
		switch family {
		case syscall.AF_INET:
			host = "0.0.0.0"
		case syscall.AF_INET6:
			host = "::"
		default:
			return Addr{}, fmt.Errorf("unknown family, %d", family)
		}
	}
	lport, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return Addr{}, err
	}
	return Addr{IP: host, Port: uint32(lport)}, nil
}

// This function only works for netstat returning addresses with a "."
// before the port (0.0.0.0.22 instead of 0.0.0.0:22).
func parseNetstatAddr(local, remote string, family uint32) (laddr, raddr Addr, err error) {
	laddr, err = parseAddr(local, family)
	if remote != "*.*" { // remote addr exists
		raddr, err = parseAddr(remote, family)
		if err != nil {
			return laddr, raddr, err
		}
	}

	return laddr, raddr, err
}

func parseNetstatUnixLine(f []string) (ConnectionStat, error) {
	if len(f) < 8 {
		return ConnectionStat{}, fmt.Errorf("wrong number of fields: expected >=8 got %d", len(f))
	}

	var netType uint32

	switch f[1] {
	case "dgram":
		netType = syscall.SOCK_DGRAM
	case "stream":
		netType = syscall.SOCK_STREAM
	default:
		return ConnectionStat{}, fmt.Errorf("unknown type: %s", f[1])
	}

	// Some Unix Socket don't have any address associated
	addr := ""
	if len(f) == 9 {
		addr = f[8]
	}

	c := ConnectionStat{
		Fd:     uint32(0), // not supported
		Family: uint32(syscall.AF_UNIX),
		Type:   uint32(netType),
		Laddr: Addr{
			IP: addr,
		},
		Status: "NONE",
		Pid:    int32(0), // not supported
	}

	return c, nil
}

// Return true if proto is the corresponding to the kind parameter
// Only for Inet lines
func hasCorrectInetProto(kind, proto string) bool {
	switch kind {
	case "all", "inet":
		return true
	case "unix":
		return false
	case "inet4":
		return !strings.HasSuffix(proto, "6")
	case "inet6":
		return strings.HasSuffix(proto, "6")
	case "tcp":
		return proto == "tcp" || proto == "tcp4" || proto == "tcp6"
	case "tcp4":
		return proto == "tcp" || proto == "tcp4"
	case "tcp6":
		return proto == "tcp6"
	case "udp":
		return proto == "udp" || proto == "udp4" || proto == "udp6"
	case "udp4":
		return proto == "udp" || proto == "udp4"
	case "udp6":
		return proto == "udp6"
	}
	return false
}

func parseNetstatA(output, kind string) ([]ConnectionStat, error) {
	var ret []ConnectionStat
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}

		switch {
		case strings.HasPrefix(fields[0], "f1"):
			// Unix lines
			if len(fields) < 2 {
				// every unix connections have two lines
				continue
			}

			c, err := parseNetstatUnixLine(fields)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Unix Address (%s): %w", line, err)
			}

			ret = append(ret, c)

		case strings.HasPrefix(fields[0], "tcp") || strings.HasPrefix(fields[0], "udp"):
			// Inet lines
			if !hasCorrectInetProto(kind, fields[0]) {
				continue
			}

			// On AIX, netstat display some connections with "*.*" as local addresses
			// Skip them as they aren't real connections.
			if fields[3] == "*.*" {
				continue
			}

			c, err := parseNetstatNetLine(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Inet Address (%s): %w", line, err)
			}

			ret = append(ret, c)
		default:
			// Header lines
			continue
		}
	}

	return ret, nil
}

func ConnectionsWithContext(ctx context.Context, kind string) ([]ConnectionStat, error) {
	args := []string{"-na"}
	switch strings.ToLower(kind) {
	default:
		fallthrough
	case "":
		kind = "all"
	case "all":
		// nothing to add
	case "inet", "inet4", "inet6":
		args = append(args, "-finet")
	case "tcp", "tcp4", "tcp6":
		args = append(args, "-finet")
	case "udp", "udp4", "udp6":
		args = append(args, "-finet")
	case "unix":
		args = append(args, "-funix")
	}

	out, err := invoke.CommandWithContext(ctx, "netstat", args...)
	if err != nil {
		return nil, err
	}

	ret, err := parseNetstatA(string(out), kind)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func ConnectionsMaxWithContext(ctx context.Context, kind string, maxConn int) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithContext(ctx, kind, 0, maxConn)
}

func ConnectionsWithoutUidsWithContext(ctx context.Context, kind string) ([]ConnectionStat, error) {
	return ConnectionsMaxWithoutUidsWithContext(ctx, kind, 0)
}

func ConnectionsMaxWithoutUidsWithContext(ctx context.Context, kind string, maxConn int) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithoutUidsWithContext(ctx, kind, 0, maxConn)
}

func ConnectionsPidWithContext(ctx context.Context, kind string, pid int32) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithContext(ctx, kind, pid, 0)
}

func ConnectionsPidWithoutUidsWithContext(ctx context.Context, kind string, pid int32) ([]ConnectionStat, error) {
	return ConnectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, 0)
}

func ConnectionsPidMaxWithContext(ctx context.Context, kind string, pid int32, maxConn int) ([]ConnectionStat, error) {
	return connectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, maxConn, false)
}

func ConnectionsPidMaxWithoutUidsWithContext(ctx context.Context, kind string, pid int32, maxConn int) ([]ConnectionStat, error) {
	return connectionsPidMaxWithoutUidsWithContext(ctx, kind, pid, maxConn, true)
}

func connectionsPidMaxWithoutUidsWithContext(ctx context.Context, _ string, pid int32, maxConn int, _ bool) ([]ConnectionStat, error) {
	// If pid is 0, return all connections
	if pid == 0 {
		return getAIXConnections(ctx, maxConn)
	}

	// For specific PID, filter connections to just that process
	return getAIXConnectionsForPid(ctx, pid, maxConn)
}

// getAIXConnections retrieves all network connections from AIX
func getAIXConnections(ctx context.Context, maxConn int) ([]ConnectionStat, error) {
	var conns []ConnectionStat

	// Get all listening sockets using netstat
	output, err := invoke.CommandWithContext(ctx, "netstat", "-Aan")
	if err != nil {
		return conns, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0

	for _, line := range lines {
		if maxConn > 0 && count >= maxConn {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Address") {
			continue
		}

		// Parse netstat output to find sockets
		// Format: Address    Family  Type      Use  Recv-Q Send-Q    Inode  Conn  Routes
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		sockAddr := fields[0]

		// Try to resolve the socket to get connection info using rmsock
		// For TCP connections
		connStat, err := resolveAIXSockToConnection(ctx, sockAddr, "tcp")
		if err == nil && connStat != nil {
			conns = append(conns, *connStat)
			count++
			continue
		}

		// Try for UDP connections
		connStat, err = resolveAIXSockToConnection(ctx, sockAddr, "udp")
		if err == nil && connStat != nil {
			conns = append(conns, *connStat)
			count++
		}
	}

	return conns, nil
}

// getAIXConnectionsForPid retrieves network connections for a specific process on AIX
func getAIXConnectionsForPid(ctx context.Context, pid int32, maxConn int) ([]ConnectionStat, error) {
	var conns []ConnectionStat

	// Get all listening sockets using netstat
	output, err := invoke.CommandWithContext(ctx, "netstat", "-Aan")
	if err != nil {
		return conns, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0

	for _, line := range lines {
		if maxConn > 0 && count >= maxConn {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Address") || strings.HasPrefix(line, "PCB/ADDR") {
			continue
		}

		// Parse netstat output: PCB/ADDR Proto Recv-Q Send-Q Local Address Foreign Address (state)
		// Example: f1000f00055cc3c0 tcp4       0      0  192.168.242.122.22    24.236.207.124.40326  ESTABLISHED
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		sockAddr := fields[0]
		proto := fields[1]
		localAddr := fields[4]
		remoteAddr := fields[5]
		state := fields[6]

		// Determine protocol type (tcp or udp)
		var protocol string
		if strings.HasPrefix(proto, "tcp") {
			protocol = "tcp"
		} else if strings.HasPrefix(proto, "udp") {
			protocol = "udp"
		} else {
			continue
		}

		// Try to resolve the socket to get PID using rmsock
		resolvedPid := resolveAIXSockToPid(ctx, sockAddr, protocol)
		if resolvedPid != pid {
			// This connection doesn't belong to our target PID
			continue
		}

		// Parse addresses
		laddr := parseAIXAddress(localAddr)
		raddr := parseAIXAddress(remoteAddr)

		// Determine socket type and family
		var socketType uint32
		var socketFamily uint32

		// Set socket type based on protocol
		if protocol == "tcp" {
			socketType = syscall.SOCK_STREAM
		} else {
			socketType = syscall.SOCK_DGRAM
		}

		// Set socket family based on proto string (tcp4, tcp6, udp4, udp6)
		if strings.HasSuffix(proto, "6") {
			socketFamily = syscall.AF_INET6
		} else {
			socketFamily = syscall.AF_INET
		}

		connStat := ConnectionStat{
			Fd:     0,
			Family: socketFamily,
			Type:   socketType,
			Laddr:  laddr,
			Raddr:  raddr,
			Status: state,
			Pid:    pid,
		}

		conns = append(conns, connStat)
		count++
	}

	return conns, nil
}

// resolveAIXSockToConnection uses AIX rmsock command to resolve a socket address to connection info
func resolveAIXSockToConnection(ctx context.Context, sockAddr string, protocol string) (*ConnectionStat, error) {
	if protocol != "tcp" && protocol != "udp" {
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	// Execute rmsock to resolve socket
	// Format for TCP: rmsock <socket_address> tcpcb
	// Format for UDP: rmsock <socket_address> inpcb
	var tcpOrUdp string
	if protocol == "tcp" {
		tcpOrUdp = "tcpcb"
	} else {
		tcpOrUdp = "inpcb"
	}

	output, err := invoke.CommandWithContext(ctx, "rmsock", sockAddr, tcpOrUdp)
	if err != nil {
		return nil, err
	}

	// Parse rmsock output to extract connection info
	outputStr := string(output)

	// Try to find PID in the output
	pid := parseAIXRmsockPid(outputStr)
	if pid == 0 {
		return nil, fmt.Errorf("could not extract PID from rmsock output")
	}

	// Build connection stat from parsed info
	connStat := &ConnectionStat{
		Fd:     0,
		Family: 0,
		Type:   0,
		Laddr:  Addr{IP: "", Port: 0},
		Raddr:  Addr{IP: "", Port: 0},
		Status: "",
		Pid:    pid,
	}

	return connStat, nil
}

// resolveAIXSockToConnectionForPid resolves socket to connection only if it matches the target PID
func resolveAIXSockToConnectionForPid(ctx context.Context, sockAddr string, protocol string, targetPid int32) (*ConnectionStat, error) {
	connStat, err := resolveAIXSockToConnection(ctx, sockAddr, protocol)
	if err != nil {
		return nil, err
	}

	if connStat == nil {
		return nil, fmt.Errorf("connection stat is nil")
	}

	if connStat.Pid != targetPid {
		return nil, fmt.Errorf("PID mismatch: expected %d, got %d", targetPid, connStat.Pid)
	}

	return connStat, nil
}

// parseAIXRmsockPid extracts PID from rmsock output
// Expected format: "The socket 0xf1000f00055be808 is being held by proccess 14287304 (sshd)."
func parseAIXRmsockPid(output string) int32 {
	// Use regex to extract PID from rmsock output
	// Pattern: "proccess <PID> ("
	re := regexp.MustCompile(`proccess\s+(\d+)\s+\(`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if pid, err := strconv.ParseInt(matches[1], 10, 32); err == nil {
			return int32(pid)
		}
	}
	return 0
}

// resolveAIXSockToPid uses rmsock to get the PID holding a socket, returns 0 if unable to resolve
func resolveAIXSockToPid(ctx context.Context, sockAddr string, protocol string) int32 {
	if protocol != "tcp" && protocol != "udp" {
		return 0
	}

	var tcpOrUdp string
	if protocol == "tcp" {
		tcpOrUdp = "tcpcb"
	} else {
		tcpOrUdp = "inpcb"
	}

	output, err := invoke.CommandWithContext(ctx, "rmsock", sockAddr, tcpOrUdp)
	// Note: rmsock may exit with status 1 even on successful resolution
	// So we try to parse the output regardless of error status

	outputStr := string(output)
	pid := parseAIXRmsockPid(outputStr)

	if pid == 0 && err != nil {
		// If we got a "Wait for exiting processes" message, it's a transient cleanup situation - skip silently
		if strings.Contains(outputStr, "Wait for exiting processes") {
			return 0
		}
		// For other errors, log debug info if we couldn't parse a PID
		// Uncomment for debugging: fmt.Fprintf(os.Stderr, "DEBUG: rmsock %s %s failed: %v, output: %s\n", sockAddr, tcpOrUdp, err, outputStr)
	}

	return pid
}

// parseAIXAddress parses an AIX address string like "192.168.242.122.22" or "24.236.207.124.40326"
// Format: IP_OCTETS separated by dots, with port as last octet(s) after the IP
func parseAIXAddress(addrStr string) Addr {
	if addrStr == "*.*" {
		return Addr{IP: "", Port: 0}
	}

	parts := strings.Split(addrStr, ".")
	if len(parts) < 2 {
		return Addr{IP: "", Port: 0}
	}

	// Last part is the port
	port := 0
	if p, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
		port = p
	}

	// Join all but last part as IP
	ip := strings.Join(parts[:len(parts)-1], ".")

	return Addr{IP: ip, Port: uint32(port)}
}
