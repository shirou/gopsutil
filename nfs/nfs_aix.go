// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package nfs

import (
	"context"
	"strconv"
	"strings"
	"sync"
)

var (
	// Cache for full stats to avoid re-parsing nfsstat
	cachedStats *StatsStat
	cacheMutex  sync.RWMutex
)

// statsWithContext returns NFS statistics from nfsstat command (internal, may be cached)
func statsWithContext(ctx context.Context) (*StatsStat, error) {
	stats := &StatsStat{
		NFSClientStats: NFSClientStat{Operations: make(map[string]uint64)},
		NFSServerStats: NFSServerStat{Operations: make(map[string]uint64)},
	}

	// Get nfsstat output
	output, err := invoke.CommandWithContext(ctx, "nfsstat")
	if err != nil {
		return nil, err
	}

	// Parse the output
	if err := parseNfsstat(string(output), stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// ClientStatsWithContext returns NFS client statistics
func ClientStatsWithContext(ctx context.Context) (*NFSClientStat, error) {
	stats, err := statsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &stats.NFSClientStats, nil
}

// ClientStats returns NFS client statistics
func ClientStats() (*NFSClientStat, error) {
	return ClientStatsWithContext(context.Background())
}

// ServerStatsWithContext returns NFS server statistics
func ServerStatsWithContext(ctx context.Context) (*NFSServerStat, error) {
	stats, err := statsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &stats.NFSServerStats, nil
}

// ServerStats returns NFS server statistics
func ServerStats() (*NFSServerStat, error) {
	return ServerStatsWithContext(context.Background())
}

// parseNfsstat parses nfsstat command output on AIX
// AIX nfsstat output is organized into sections:
//   - Server rpc: RPC statistics for NFS server
//   - Server nfs: NFS operation statistics for server
//   - Client rpc: RPC statistics for NFS client
//   - Client nfs: NFS operation statistics for client
//
// Each RPC section has subsections for "Connection oriented" (TCP) and "Connectionless" (UDP)
// NFS sections report per-operation statistics broken down by NFSv2 and NFSv3
func parseNfsstat(output string, stats *StatsStat) error {
	lines := strings.Split(output, "\n")

	var section string       // Track current section: "server_rpc", "server_nfs", "client_rpc", "client_nfs"
	var transportMode string // Track "Connection oriented" (TCP) vs "Connectionless" (UDP)
	var nfsVersion string    // Track "Version 2" vs "Version 3"

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		// Detect sections
		if strings.Contains(line, "Server rpc:") {
			section = "server_rpc"
			continue
		}
		if strings.Contains(line, "Server nfs:") {
			section = "server_nfs"
			continue
		}
		if strings.Contains(line, "Client rpc:") {
			section = "client_rpc"
			continue
		}
		if strings.Contains(line, "Client nfs:") {
			section = "client_nfs"
			continue
		}

		// Detect transport mode
		if strings.Contains(line, "Connection oriented") {
			transportMode = "connection_oriented"
			continue
		}
		if strings.Contains(line, "Connectionless") {
			transportMode = "connectionless"
			continue
		}

		// Detect NFS version
		if strings.Contains(line, "Version 2:") {
			nfsVersion = "v2"
			continue
		}
		if strings.Contains(line, "Version 3:") {
			nfsVersion = "v3"
			continue
		}

		// Parse data lines based on section
		switch section {
		case "server_rpc":
			if err := parseRPCServerLine(line, transportMode, &stats.RPCServerStats); err != nil {
				continue
			}
		case "server_nfs":
			if err := parseNFSServerLine(line, nfsVersion, &stats.NFSServerStats); err != nil {
				continue
			}
		case "client_rpc":
			if err := parseRPCClientLine(line, transportMode, &stats.RPCClientStats); err != nil {
				continue
			}
		case "client_nfs":
			if err := parseNFSClientLine(line, nfsVersion, &stats.NFSClientStats); err != nil {
				continue
			}
		}
	}

	return nil
}

// parseRPCServerLine parses an RPC server statistics line from AIX nfsstat
// AIX reports separate metrics for "Connection oriented" (TCP) and "Connectionless" (UDP) transports
// Both have: calls badcalls nullrecv badlen xdrcall dupchecks dupreqs
func parseRPCServerLine(line string, transportMode string, stats *RPCServerStat) error {
	// Skip header lines
	if strings.Contains(line, "calls") && strings.Contains(line, "badcalls") {
		return nil
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}

	// Parse based on transport mode and field count
	if transportMode == "connection_oriented" {
		if len(fields) >= 7 {
			// Connection oriented: calls badcalls nullrecv badlen xdrcall dupchecks dupreqs
			stats.Calls += parseUint64(fields[0])
			stats.BadCalls += parseUint64(fields[1])
			stats.NullRecv += parseUint64(fields[2])
			stats.BadLen += parseUint64(fields[3])
			stats.XdrCall += parseUint64(fields[4])
			stats.DupChecks += parseUint64(fields[5])
			stats.DupReqs += parseUint64(fields[6])
		}
	} else if transportMode == "connectionless" {
		if len(fields) >= 7 {
			// Connectionless: calls badcalls nullrecv badlen xdrcall dupchecks dupreqs
			stats.Calls += parseUint64(fields[0])
			stats.BadCalls += parseUint64(fields[1])
			stats.NullRecv += parseUint64(fields[2])
			stats.BadLen += parseUint64(fields[3])
			stats.XdrCall += parseUint64(fields[4])
			stats.DupChecks += parseUint64(fields[5])
			stats.DupReqs += parseUint64(fields[6])
		}
	}

	return nil
}

// parseRPCClientLine parses an RPC client statistics line from AIX nfsstat
// AIX reports separate metrics for "Connection oriented" (TCP) and "Connectionless" (UDP) transports
// Connection oriented has: calls badcalls badxids timeouts newcreds badverfs timers nomem cantconn interrupts
// Connectionless has: calls badcalls retrans badxids timeouts newcreds badverfs timers nomem cantsend
func parseRPCClientLine(line string, transportMode string, stats *RPCClientStat) error {
	// Skip header lines
	if strings.Contains(line, "calls") {
		return nil
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}

	if transportMode == "connection_oriented" {
		// Connection oriented: calls badcalls badxids timeouts newcreds badverfs timers
		if len(fields) >= 7 {
			stats.Calls += parseUint64(fields[0])
			stats.BadCalls += parseUint64(fields[1])
			stats.BadXIDs += parseUint64(fields[2])
			stats.Timeouts += parseUint64(fields[3])
			stats.NewCreds += parseUint64(fields[4])
			stats.BadVerfs += parseUint64(fields[5])
			stats.Timers += parseUint64(fields[6])
		}
		// nomem cantconn interrupts (on next iteration for these fields)
		if len(fields) >= 3 && (strings.HasPrefix(line, "nomem") || strings.HasPrefix(line, "0")) {
			if strings.HasPrefix(line, "nomem") || (len(fields) == 3 && fields[0] != "nomem") {
				return nil // Skip this parse, it's a different line
			}
		}
	} else if transportMode == "connectionless" {
		// Connectionless: calls badcalls retrans badxids timeouts newcreds badverfs
		if len(fields) >= 7 {
			stats.Calls += parseUint64(fields[0])
			stats.BadCalls += parseUint64(fields[1])
			stats.Retrans += parseUint64(fields[2])
			stats.BadXIDs += parseUint64(fields[3])
			stats.Timeouts += parseUint64(fields[4])
			stats.NewCreds += parseUint64(fields[5])
			stats.BadVerfs += parseUint64(fields[6])
		}
		// timers nomem cantsend (on next iteration)
		if len(fields) == 3 && strings.HasPrefix(line, "0") {
			stats.Timers += parseUint64(fields[0])
			stats.NoMem += parseUint64(fields[1])
			stats.CantSend += parseUint64(fields[2])
		}
	}

	// Parse nomem/cantconn/interrupts/timers lines (3 fields)
	if len(fields) == 3 {
		// Check if this is a continuation line
		first, _ := strconv.ParseUint(fields[0], 10, 64)
		_ = first // Use the parsed value to detect numeric line
		if strings.HasPrefix(line, "nomem") {
			stats.NoMem += parseUint64(fields[0])
			if transportMode == "connection_oriented" {
				stats.CantConn += parseUint64(fields[1])
			} else {
				stats.Timers += parseUint64(fields[1])
			}
			stats.Interrupts += parseUint64(fields[2])
		} else if strings.HasPrefix(line, "timers") {
			stats.Timers += parseUint64(fields[0])
			stats.NoMem += parseUint64(fields[1])
			stats.CantSend += parseUint64(fields[2])
		}
	}

	return nil
}

// parseNFSServerLine parses an NFS server statistics line from AIX nfsstat
// AIX reports per-operation statistics broken down by NFSv2 and NFSv3
// Format: "calls badcalls public_v2 public_v3" followed by per-operation stats with percentages
func parseNFSServerLine(line string, nfsVersion string, stats *NFSServerStat) error {
	// Parse "Server nfs:" header line
	if strings.Contains(line, "calls") && strings.Contains(line, "badcalls") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "calls" {
			// This is the main header, skip
			return nil
		}
		// This might be the calls/badcalls/public_v2/public_v3 line
		if len(fields) >= 4 {
			stats.Calls = parseUint64(fields[0])
			stats.BadCalls = parseUint64(fields[1])
			if strings.HasPrefix(fields[0], "0") && len(fields) >= 4 {
				stats.Calls = parseUint64(fields[0])
				stats.BadCalls = parseUint64(fields[1])
				if len(fields) > 2 {
					stats.PublicV2 = parseUint64(fields[2])
				}
				if len(fields) > 3 {
					stats.PublicV3 = parseUint64(fields[3])
				}
			}
		}
		return nil
	}

	// Parse operation lines (contains percentage)
	if strings.Contains(line, "%") {
		return parseNFSOperationLine(line, nfsVersion, stats.Operations)
	}

	// Parse the initial calls/badcalls/public_v2/public_v3 line
	fields := strings.Fields(line)
	if len(fields) >= 4 && !strings.Contains(line, "%") && !strings.Contains(line, "Version") && !strings.Contains(line, "null") {
		// Could be: 0 0 0 0 (calls badcalls public_v2 public_v3)
		if _, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
			stats.Calls = parseUint64(fields[0])
			stats.BadCalls = parseUint64(fields[1])
			if len(fields) > 2 {
				stats.PublicV2 = parseUint64(fields[2])
			}
			if len(fields) > 3 {
				stats.PublicV3 = parseUint64(fields[3])
			}
		}
	}

	return nil
}

// parseNFSClientLine parses an NFS client statistics line from AIX nfsstat
// AIX reports per-operation statistics broken down by NFSv2 and NFSv3
// Format: "calls badcalls clgets cltoomany" followed by per-operation stats with percentages
func parseNFSClientLine(line string, nfsVersion string, stats *NFSClientStat) error {
	// Parse header line: calls badcalls clgets cltoomany
	if strings.HasPrefix(line, "calls") && strings.Contains(line, "badcalls") {
		return nil
	}

	// Parse operation lines (contains percentage)
	if strings.Contains(line, "%") {
		return parseNFSOperationLine(line, nfsVersion, stats.Operations)
	}

	// Parse the calls/badcalls/clgets/cltoomany line
	fields := strings.Fields(line)
	if len(fields) >= 4 && !strings.Contains(line, "%") && !strings.Contains(line, "Version") {
		if _, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
			stats.Calls = parseUint64(fields[0])
			stats.BadCalls = parseUint64(fields[1])
			if len(fields) > 2 {
				stats.ClGets = parseUint64(fields[2])
			}
			if len(fields) > 3 {
				stats.ClTooMany = parseUint64(fields[3])
			}
		}
	}

	return nil
}

// parseNFSOperationLine parses NFS operation statistics line from AIX nfsstat
// AIX output format: operation1 count1% operation2 count2% ...
// Example: "null       0 0%       getattr    0 0%       setattr    0 0%"
// Operations are specific to the NFS version context (v2 or v3)
func parseNFSOperationLine(line string, nfsVersion string, operations map[string]uint64) error {
	// Format: operation1 count1% operation2 count2% ...
	// Split on whitespace and process pairs of (operation, count)
	fields := strings.Fields(line)

	for i := 0; i < len(fields)-1; i += 2 {
		opName := fields[i]
		countStr := fields[i+1]

		// Remove percentage sign if present
		countStr = strings.TrimSuffix(countStr, "%")

		// Parse count
		count := parseUint64(countStr)

		// Determine version prefix if needed
		key := opName
		if nfsVersion != "" {
			key = nfsVersion + "_" + opName
		}

		operations[key] = count
	}

	return nil
}

// parseUint64 safely parses a string to uint64, returning 0 on error
func parseUint64(s string) uint64 {
	val, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return val
}
