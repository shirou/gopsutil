// SPDX-License-Identifier: BSD-3-Clause
//go:build aix

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/nfs"
	"github.com/shirou/gopsutil/v4/process"
)

func main() {
	ctx := context.Background()
	outfile := "gopsutil_metrics.txt"

	f, err := os.Create(outfile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	write := func(format string, args ...interface{}) {
		fmt.Fprintf(f, format+"\n", args...)
		fmt.Printf(format+"\n", args...)
	}

	// Helper to convert socket type to readable string
	socketTypeStr := func(sockType uint32) string {
		switch sockType {
		case syscall.SOCK_STREAM:
			return "TCP"
		case syscall.SOCK_DGRAM:
			return "UDP"
		default:
			return fmt.Sprintf("UNKNOWN(%d)", sockType)
		}
	}

	write("================================================================================")
	write("GOPSUTIL AIX METRICS TEST - %s", time.Now().Format(time.RFC3339))
	write("================================================================================\n")

	// ============================================================================
	// HOST INFORMATION
	// ============================================================================
	write("--- HOST INFORMATION ---")
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		write("ERROR getting host info: %v", err)
	} else {
		write("Hostname: %s", hostInfo.Hostname)
		write("Uptime: %d seconds", hostInfo.Uptime)
		write("Boot Time: %d", hostInfo.BootTime)
		write("Procs: %d", hostInfo.Procs)
		write("OS: %s", hostInfo.OS)
		write("Platform: %s", hostInfo.Platform)
		write("Platform Family: %s", hostInfo.PlatformFamily)
		write("Platform Version: %s", hostInfo.PlatformVersion)
		write("Kernel Version: %s", hostInfo.KernelVersion)
		write("Kernel Arch: %s", hostInfo.KernelArch)
		write("Virtualization System: %s", hostInfo.VirtualizationSystem)
		write("Virtualization Role: %s", hostInfo.VirtualizationRole)
		write("Host ID (UUID): %s", hostInfo.HostID)
	}
	write("")

	// ============================================================================
	// CPU INFORMATION
	// ============================================================================
	write("--- CPU INFORMATION ---")
	cpuCount, err := cpu.Counts(false)
	if err != nil {
		write("ERROR getting CPU count: %v", err)
	} else {
		write("Logical CPU Count: %d", cpuCount)
	}

	cpuPhysical, err := cpu.Counts(true)
	if err != nil {
		write("ERROR getting physical CPU count: %v", err)
	} else {
		write("Physical CPU Count: %d", cpuPhysical)
	}

	cpuInfos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		write("ERROR getting CPU info: %v", err)
	} else {
		write("CPU Count from Info: %d", len(cpuInfos))
		for i, info := range cpuInfos {
			write("  CPU %d: %s (Family: %s, Model: %s, Stepping: %d, Cores: %d, MHz: %.0f)",
				i, info.ModelName, info.Family, info.Model, info.Stepping, info.Cores, info.Mhz)
		}
	}

	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		write("ERROR getting CPU times: %v", err)
	} else {
		write("CPU Times:")
		for _, t := range times {
			write("  CPU: %s | User: %.2f | System: %.2f | Idle: %.2f | Nice: %.2f | Iowait: %.2f | Irq: %.2f | Softirq: %.2f | Steal: %.2f | Guest: %.2f | GuestNice: %.2f",
				t.CPU, t.User, t.System, t.Idle, t.Nice, t.Iowait, t.Irq, t.Softirq, t.Steal, t.Guest, t.GuestNice)
		}
	}

	percent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		write("ERROR getting CPU percent: %v", err)
	} else {
		write("CPU Percent: %v", percent)
	}
	write("")

	// ============================================================================
	// MEMORY INFORMATION
	// ============================================================================
	write("--- MEMORY INFORMATION ---")
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		write("ERROR getting virtual memory: %v", err)
	} else {
		write("Virtual Memory:")
		write("  Total: %d MB", memInfo.Total/1024/1024)
		write("  Available: %d MB", memInfo.Available/1024/1024)
		write("  Used: %d MB", memInfo.Used/1024/1024)
		write("  Used Percent: %.2f%%", memInfo.UsedPercent)
		write("  Free: %d MB", memInfo.Free/1024/1024)
		write("  Buffers: %d MB", memInfo.Buffers/1024/1024)
		write("  Cached: %d MB", memInfo.Cached/1024/1024)
		write("  Shared: %d MB", memInfo.Shared/1024/1024)
	}

	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		write("ERROR getting swap memory: %v", err)
	} else {
		write("Swap Memory:")
		write("  Total: %d MB", swapInfo.Total/1024/1024)
		write("  Used: %d MB", swapInfo.Used/1024/1024)
		write("  Free: %d MB", swapInfo.Free/1024/1024)
		write("  Used Percent: %.2f%%", swapInfo.UsedPercent)
	}
	write("")

	// ============================================================================
	// LOAD AVERAGE
	// ============================================================================
	write("--- LOAD AVERAGE ---")
	loadAvg, err := load.AvgWithContext(ctx)
	if err != nil {
		write("ERROR getting load average: %v", err)
	} else {
		write("Load Average: 1m=%.2f, 5m=%.2f, 15m=%.2f", loadAvg.Load1, loadAvg.Load5, loadAvg.Load15)
	}

	loadMisc, err := load.MiscWithContext(ctx)
	if err != nil {
		write("ERROR getting load misc: %v", err)
	} else {
		write("Load Misc: Running Processes=%d, Blocked Processes=%d", loadMisc.ProcsRunning, loadMisc.ProcsBlocked)
	}
	write("")

	// ============================================================================
	// DISK INFORMATION
	// ============================================================================
	write("--- DISK INFORMATION ---")
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		write("ERROR getting disk partitions: %v", err)
	} else {
		write("Disk Partitions: %d", len(partitions))
		for _, part := range partitions {
			write("  Device: %s | Mountpoint: %s | FS: %s | Opts: %s",
				part.Device, part.Mountpoint, part.Fstype, part.Opts)
		}
	}

	for _, part := range partitions {
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err != nil {
			write("  ERROR getting usage for %s: %v", part.Mountpoint, err)
			continue
		}
		write("  %s: Total=%d MB, Used=%d MB, Free=%d MB, UsedPercent=%.2f%%",
			part.Mountpoint, usage.Total/1024/1024, usage.Used/1024/1024, usage.Free/1024/1024, usage.UsedPercent)
	}
	write("")

	// ============================================================================
	// NETWORK INFORMATION
	// ============================================================================
	write("--- NETWORK INFORMATION ---")
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		write("ERROR getting network interfaces: %v", err)
	} else {
		write("Network Interfaces: %d", len(interfaces))
		for _, iface := range interfaces {
			write("  Interface: %s", iface.Name)
			write("    MTU: %d", iface.MTU)
			for _, addr := range iface.Addrs {
				write("    Address: %s", addr.Addr)
			}
		}
	}

	ioCounters, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		write("ERROR getting network IO counters: %v", err)
	} else {
		write("Network IO Counters: %d interfaces", len(ioCounters))
		for _, counter := range ioCounters {
			write("  %s: BytesRecv=%d, BytesSent=%d, PacketsRecv=%d, PacketsSent=%d, Errin=%d, Errout=%d, Dropin=%d, Dropout=%d",
				counter.Name, counter.BytesRecv, counter.BytesSent, counter.PacketsRecv, counter.PacketsSent,
				counter.Errin, counter.Errout, counter.Dropin, counter.Dropout)
		}
	}

	// Per-process connections
	currentPid := os.Getpid()
	connWithPid, err := net.ConnectionsPidWithContext(ctx, "all", int32(currentPid))
	if err != nil {
		write("ERROR getting connections for current PID: %v", err)
	} else {
		write("Connections for PID %d: %d connections", currentPid, len(connWithPid))
		for i, conn := range connWithPid {
			if i >= 5 { // Show first 5
				write("  ... and %d more connections", len(connWithPid)-5)
				break
			}
			write("  %s %s:%d -> %s:%d (Status: %s)", conn.Type, conn.Laddr.IP, conn.Laddr.Port, conn.Raddr.IP, conn.Raddr.Port, conn.Status)
		}
	}
	write("")

	// ============================================================================
	// NFS INFORMATION
	// ============================================================================
	write("--- NFS INFORMATION ---")
	clientStats, err := nfs.ClientStatsWithContext(ctx)
	if err != nil {
		write("ERROR getting NFS client stats: %v", err)
	} else {
		write("NFS Client Stats:")
		write("  Calls: %d", clientStats.Calls)
		write("  Bad Calls: %d", clientStats.BadCalls)
		write("  ClGets: %d", clientStats.ClGets)
		write("  ClTooMany: %d", clientStats.ClTooMany)
		write("  Operations: %d types", len(clientStats.Operations))
		for op, count := range clientStats.Operations {
			write("    %s: %d", op, count)
		}
	}

	serverStats, err := nfs.ServerStatsWithContext(ctx)
	if err != nil {
		write("ERROR getting NFS server stats: %v", err)
	} else {
		write("NFS Server Stats:")
		write("  Calls: %d", serverStats.Calls)
		write("  Bad Calls: %d", serverStats.BadCalls)
		write("  PublicV2: %d", serverStats.PublicV2)
		write("  PublicV3: %d", serverStats.PublicV3)
		write("  Operations: %d types", len(serverStats.Operations))
		for op, count := range serverStats.Operations {
			write("    %s: %d", op, count)
		}
	}
	write("")

	// ============================================================================
	// PROCESS INFORMATION
	// ============================================================================
	write("--- PROCESS INFORMATION ---")

	// Get current process
	currentProc, err := process.NewProcess(int32(currentPid))
	if err != nil {
		write("ERROR getting current process: %v", err)
	} else {
		write("Current Process (PID=%d):", currentPid)
		if name, err := currentProc.NameWithContext(ctx); err == nil {
			write("  Name: %s", name)
		}
		if ppid, err := currentProc.PpidWithContext(ctx); err == nil {
			write("  Parent PID: %d", ppid)
		}
		if status, err := currentProc.StatusWithContext(ctx); err == nil {
			write("  Status: %v", status)
		}
		if cmdline, err := currentProc.CmdlineWithContext(ctx); err == nil {
			write("  Cmdline: %s", cmdline)
		}
		if cmdlineSlice, err := currentProc.CmdlineSliceWithContext(ctx); err == nil {
			write("  Cmdline Slice: %v", cmdlineSlice)
		}
		if exe, err := currentProc.ExeWithContext(ctx); err == nil {
			write("  Exe: %s", exe)
		}
		if cwd, err := currentProc.CwdWithContext(ctx); err == nil {
			write("  CWD: %s", cwd)
		}
		if uids, err := currentProc.UidsWithContext(ctx); err == nil {
			write("  UIDs: %v", uids)
		}
		if gids, err := currentProc.GidsWithContext(ctx); err == nil {
			write("  GIDs: %v", gids)
		}
		if groups, err := currentProc.GroupsWithContext(ctx); err == nil {
			write("  Groups: %v", groups)
		}
		if memInfo, err := currentProc.MemoryInfoWithContext(ctx); err == nil {
			write("  Memory Info: RSS=%d MB, VMS=%d MB", memInfo.RSS/1024/1024, memInfo.VMS/1024/1024)
		}
		if memInfoEx, err := currentProc.MemoryInfoExWithContext(ctx); err == nil {
			write("  Memory Info Ex: RSS=%d MB, VMS=%d MB, Shared=%d MB",
				memInfoEx.RSS/1024/1024, memInfoEx.VMS/1024/1024, memInfoEx.Shared/1024/1024)
		}
		if cpuTimes, err := currentProc.TimesWithContext(ctx); err == nil {
			write("  CPU Times: User=%.2f, System=%.2f, Iowait=%.2f", cpuTimes.User, cpuTimes.System, cpuTimes.Iowait)
		}
		if cpuAff, err := currentProc.CPUAffinityWithContext(ctx); err == nil {
			write("  CPU Affinity: %v", cpuAff)
		}
		if nice, err := currentProc.NiceWithContext(ctx); err == nil {
			write("  Nice: %d", nice)
		}
		if numThreads, err := currentProc.NumThreadsWithContext(ctx); err == nil {
			write("  Num Threads: %d", numThreads)
		}
		if numFds, err := currentProc.NumFDsWithContext(ctx); err == nil {
			write("  Num FDs: %d", numFds)
		}
		if pageFaults, err := currentProc.PageFaultsWithContext(ctx); err == nil && pageFaults != nil {
			write("  Page Faults: MinorFaults=%d, MajorFaults=%d, ChildMinorFaults=%d, ChildMajorFaults=%d",
				pageFaults.MinorFaults, pageFaults.MajorFaults, pageFaults.ChildMinorFaults, pageFaults.ChildMajorFaults)
		} else if err != nil {
			write("  Page Faults: Error - %v", err)
		}
		if threads, err := currentProc.ThreadsWithContext(ctx); err == nil {
			write("  Threads: %d", len(threads))
		} else if err != nil {
			write("  Threads: Error - %v", err)
		}
		if connections, err := currentProc.ConnectionsWithContext(ctx); err == nil {
			write("  Connections: %d", len(connections))
		} else if err != nil {
			write("  Connections: Error - %v", err)
		}
		if numCtxSwitches, err := currentProc.NumCtxSwitchesWithContext(ctx); err == nil && numCtxSwitches != nil {
			write("  Context Switches: Voluntary=%d, Involuntary=%d",
				numCtxSwitches.Voluntary, numCtxSwitches.Involuntary)
		} else if err != nil {
			write("  Context Switches: Error - %v", err)
		}
		if ioCounters, err := currentProc.IOCountersWithContext(ctx); err == nil && ioCounters != nil {
			write("  IO Counters: ReadCount=%d, WriteCount=%d, ReadBytes=%d, WriteBytes=%d",
				ioCounters.ReadCount, ioCounters.WriteCount, ioCounters.ReadBytes, ioCounters.WriteBytes)
		} else if err != nil {
			write("  IO Counters: Error - %v", err)
		}
		if memMaps, err := currentProc.MemoryMapsWithContext(ctx, false); err == nil && memMaps != nil {
			write("  Memory Maps: %d regions", len(*memMaps))
			for i, mapStat := range *memMaps {
				if i >= 10 { // Show first 10 memory maps
					write("    ... and %d more regions", len(*memMaps)-10)
					break
				}
				write("    Region %d: Path=%s, Size=%d bytes, RSS=%d bytes", i+1, mapStat.Path, mapStat.Size, mapStat.Rss)
			}
		} else if err != nil {
			write("  Memory Maps: Error - %v", err)
		}
	}
	write("")

	// ============================================================================
	// SSHD PROCESS CONNECTIONS TEST
	// ============================================================================
	write("--- SSHD PROCESS CONNECTIONS ---")
	allProcs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		write("ERROR getting all processes: %v", err)
	} else {
		// Find all sshd processes (there may be multiple: listener + session handlers)
		var sshdProcs []*process.Process
		for _, proc := range allProcs {
			name, err := proc.NameWithContext(ctx)
			if err == nil && name == "sshd" {
				sshdProcs = append(sshdProcs, proc)
			}
		}

		if len(sshdProcs) == 0 {
			write("sshd process not found")
		} else {
			write("Found %d sshd processes", len(sshdProcs))
			for _, sshdProc := range sshdProcs {
				write("  sshd (PID=%d):", sshdProc.Pid)
				conns, err := sshdProc.ConnectionsWithContext(ctx)
				if err != nil {
					write("    ERROR getting connections: %v", err)
				} else {
					write("    Connections: %d", len(conns))
					for i, conn := range conns {
						if i >= 5 { // Show first 5 connections per process
							write("      ... and %d more connections", len(conns)-5)
							break
						}
						write("      %s: %s:%d <-> %s:%d (Status: %s)",
							socketTypeStr(conn.Type), conn.Laddr.IP, conn.Laddr.Port, conn.Raddr.IP, conn.Raddr.Port,
							conn.Status)
					}
				}
			}
		}
	}
	write("")

	// Get all processes
	write("--- ALL PROCESSES ---")
	allProcs, err = process.ProcessesWithContext(ctx)
	if err != nil {
		write("ERROR getting all processes: %v", err)
	} else {
		write("Total Processes: %d", len(allProcs))
		write("First 20 processes:")
		count := 0
		for _, proc := range allProcs {
			if count >= 20 {
				break
			}
			name, _ := proc.NameWithContext(ctx)
			status, _ := proc.StatusWithContext(ctx)
			write("  PID=%d, Name=%s, Status=%v", proc.Pid, name, status)
			count++
		}
	}
	write("")

	write("================================================================================")
	write("END OF METRICS - Output written to: %s", outfile)
	write("================================================================================")

	log.Printf("Metrics output written to: %s", outfile)
}
