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

// netProtocols is the default set of protocols returned when no filter is specified.
var netProtocols = []string{
	"ip",
	"icmp",
	"tcp",
	"udp",
}

func ProtoCountersWithContext(ctx context.Context, protocols []string) ([]ProtoCountersStat, error) {
	out, err := invoke.CommandWithContext(ctx, "netstat", "-s")
	if err != nil {
		return nil, err
	}
	if len(protocols) == 0 {
		protocols = netProtocols
	}
	return parseNetstatS(string(out), protocols)
}

// parseNetstatS parses AIX netstat -s output into per-protocol statistics.
// Only protocols present in the protocols filter are returned.
func parseNetstatS(output string, protocols []string) ([]ProtoCountersStat, error) {
	wantAll := len(protocols) == 0
	want := make(map[string]bool)
	for _, p := range protocols {
		want[strings.ToLower(p)] = true
	}

	// Split output into protocol sections. Section headers are "<proto>:" at column 0.
	sections := make(map[string][]string)
	var currentProto string
	for _, line := range strings.Split(output, "\n") {
		if line != "" && line[0] != '\t' && line[0] != ' ' && strings.HasSuffix(strings.TrimSpace(line), ":") {
			currentProto = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(line), ":"))
			continue
		}
		if currentProto != "" {
			sections[currentProto] = append(sections[currentProto], line)
		}
	}

	var ret []ProtoCountersStat
	for proto, lines := range sections {
		if !wantAll && !want[proto] {
			continue
		}

		stats := make(map[string]int64)
		for _, line := range lines {
			// Only parse top-level stats (single-tab prefix).
			// Skip sub-items (double-tab or deeper) to avoid double-counting.
			if !strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "\t\t") {
				continue
			}

			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			// Extract leading number. Format: "<number> <description>"
			// Some lines have parenthetical data: "3958645 data packets (3187866062 bytes)"
			idx := strings.IndexByte(trimmed, ' ')
			if idx < 0 {
				continue
			}
			val, err := strconv.ParseInt(trimmed[:idx], 10, 64)
			if err != nil {
				continue
			}
			desc := strings.TrimSpace(trimmed[idx+1:])
			key := descToCamelCase(desc)
			if key != "" {
				stats[key] = val
			}
		}

		if len(stats) > 0 {
			ret = append(ret, ProtoCountersStat{
				Protocol: proto,
				Stats:    stats,
			})
		}
	}
	return ret, nil
}

// descToCamelCase converts a netstat -s description to a camelCase key.
// e.g. "packets sent" -> "packetsSent", "total packets received" -> "totalPacketsReceived"
// Parenthetical suffixes are stripped: "data packets (3187866062 bytes)" -> "dataPackets"
// Non-alphanumeric characters are treated as word boundaries: "ack-only" -> "ackOnly", "icmp_error" -> "icmpError"
func descToCamelCase(desc string) string {
	// Strip parenthetical suffix
	if idx := strings.IndexByte(desc, '('); idx > 0 {
		desc = strings.TrimSpace(desc[:idx])
	}

	// Replace non-alphanumeric characters (except spaces) with spaces so they
	// act as word boundaries. This handles hyphens, underscores, slashes,
	// angle brackets, and other punctuation found in netstat descriptions.
	cleaned := make([]byte, len(desc))
	for i := 0; i < len(desc); i++ {
		c := desc[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == ' ' {
			cleaned[i] = c
		} else {
			cleaned[i] = ' '
		}
	}
	desc = string(cleaned)

	words := strings.Fields(desc)
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	for i, w := range words {
		w = strings.ToLower(w)
		if w == "" {
			continue
		}
		if i == 0 {
			b.WriteString(w)
		} else {
			b.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	return b.String()
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

// netstatAanEntry represents a parsed line from `netstat -Aan` output,
// pairing a socket control block address with the parsed connection info.
type netstatAanEntry struct {
	sockAddr string         // hex socket address from netstat (e.g. "f1000f00002d8bc0")
	conn     ConnectionStat // parsed connection details
	proto    string         // raw protocol string (tcp, tcp4, tcp6, udp, etc.)
}

// parseNetstatAan parses `netstat -Aan` output which prepends a socket address
// to each connection line. For inet lines, the socket address is stripped before
// passing to parseNetstatNetLine. For unix lines, the full fields (including
// socket address) are passed to parseNetstatUnixLine, which expects the socket
// address at index 0.
//
// Inet line format:  <sockAddr> <proto> <recvq> <sendq> <laddr> <raddr> [<state>]
// Unix line format:  <sockAddr> <type> <recvq> <sendq> <inode> <conn> <refs> <nextref> [<addr>]
// Unix sockets span two lines; the second line (single hex field) is skipped.
func parseNetstatAan(output, kind string) ([]netstatAanEntry, error) {
	var ret []netstatAanEntry
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch {
		case strings.HasPrefix(fields[1], "tcp") || strings.HasPrefix(fields[1], "udp"):
			// Inet line
			if !hasCorrectInetProto(kind, fields[1]) {
				continue
			}

			if len(fields) < 6 {
				continue
			}

			// Skip connections with "*.*" as local address (no port bound)
			if fields[4] == "*.*" {
				continue
			}

			// Strip the socket address and parse the remaining fields
			connLine := strings.Join(fields[1:], " ")
			conn, err := parseNetstatNetLine(connLine)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Inet Address (%s): %w", line, err)
			}

			ret = append(ret, netstatAanEntry{
				sockAddr: fields[0],
				conn:     conn,
				proto:    fields[1],
			})

		case fields[1] == "dgram" || fields[1] == "stream":
			// Unix socket line
			if kind != "all" && kind != "unix" {
				continue
			}

			conn, err := parseNetstatUnixLine(fields)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Unix Address (%s): %w", line, err)
			}

			ret = append(ret, netstatAanEntry{
				sockAddr: fields[0],
				conn:     conn,
				proto:    "unix",
			})

		default:
			// Header lines, section separators, unix continuation lines
			continue
		}
	}

	return ret, nil
}

func connectionsPidMaxWithoutUidsWithContext(ctx context.Context, kind string, pid int32, maxConn int, _ bool) ([]ConnectionStat, error) {
	// Normalize kind
	kind = strings.ToLower(kind)
	switch kind {
	case "all", "inet", "inet4", "inet6", "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "unix":
		// valid
	default:
		kind = "all"
	}

	// Build netstat args with -A to include socket addresses for PID resolution
	args := []string{"-Aan"}
	switch kind {
	case "inet", "inet4", "inet6", "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		args = append(args, "-finet")
	case "unix":
		args = append(args, "-funix")
	}

	out, err := invoke.CommandWithContext(ctx, "netstat", args...)
	if err != nil {
		return nil, err
	}

	entries, err := parseNetstatAan(string(out), kind)
	if err != nil {
		return nil, err
	}

	var conns []ConnectionStat
	for i := range entries {
		if maxConn > 0 && len(conns) >= maxConn {
			break
		}

		// Resolve PID via rmsock for inet connections
		var connPid int32
		if entries[i].proto != "unix" {
			var protocol string
			if strings.HasPrefix(entries[i].proto, "tcp") {
				protocol = "tcp"
			} else {
				protocol = "udp"
			}
			connPid = resolveAIXSockToPid(ctx, entries[i].sockAddr, protocol)
		}

		// If filtering by PID, skip non-matching connections
		if pid > 0 && connPid != pid {
			continue
		}

		entries[i].conn.Pid = connPid
		conns = append(conns, entries[i].conn)
	}

	return conns, nil
}

// rmsockPidRe matches PID in rmsock output. AIX rmsock has a known typo
// spelling "process" as "proc" + "cess" (double c), so we match both
// spellings with proc+ess (one or more 'c').
// Expected output: "The socket 0x... is being held by proc" + "cess 14287304 (sshd)."
var rmsockPidRe = regexp.MustCompile(`proc+ess\s+(\d+)\s+\(`)

// parseAIXRmsockPid extracts PID from rmsock output.
func parseAIXRmsockPid(output string) int32 {
	matches := rmsockPidRe.FindStringSubmatch(output)
	if len(matches) > 1 {
		if pid, err := strconv.ParseInt(matches[1], 10, 32); err == nil {
			return int32(pid)
		}
	}
	return 0
}

// resolveAIXSockToPid uses rmsock to get the PID holding a socket, returns 0 if unable to resolve.
func resolveAIXSockToPid(ctx context.Context, sockAddr, protocol string) int32 {
	if protocol != "tcp" && protocol != "udp" {
		return 0
	}

	var cbType string
	if protocol == "tcp" {
		cbType = "tcpcb"
	} else {
		cbType = "inpcb"
	}

	output, err := invoke.CommandWithContext(ctx, "rmsock", sockAddr, cbType)
	// rmsock exits with status 1 even on successful resolution,
	// so we parse the output regardless of error status.

	outputStr := string(output)
	pid := parseAIXRmsockPid(outputStr)

	if pid == 0 && err != nil {
		// "Wait for exiting processes" is a transient cleanup situation - skip silently
		if strings.Contains(outputStr, "Wait for exiting processes") {
			return 0
		}
	}

	return pid
}
