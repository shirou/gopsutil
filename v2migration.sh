# This script is a helper of migration to gopsutil v2 using gorename
# 
# go get golang.org/x/tools/cmd/gorename

#
# Note:
#   process has IOCounters() for file IO, and also NetIOCounters() for Net IO.
#   This scripts replace process.NetIOCounters() to IOCounters().
#   So you need hand-fixing process.


TARGETS=`cat <<EOF
CPUTimesStat -> TimesStat
CPUInfoStat -> InfoStat
CPUTimes -> Times
CPUInfo -> Info
CPUCounts -> Counts
CPUPercent -> Percent
DiskUsageStat -> UsageStat
DiskPartitionStat -> PartitionStat
DiskIOCountersStat -> IOCountersStat
DiskPartitions -> Partitions
DiskIOCounters -> IOCounters
DiskUsage -> Usage
HostInfoStat -> InfoStat
HostInfo -> Info
GetVirtualization -> Virtualization
GetPlatformInformation -> PlatformInformation
LoadAvgStat -> AvgStat
LoadAvg -> Avg
NetIOCountersStat -> IOCountersStat
NetConnectionStat -> ConnectionStat
NetProtoCountersStat -> ProtoCountersStat
NetInterfaceAddr -> InterfaceAddr
NetInterfaceStat -> InterfaceStat
NetFilterStat -> FilterStat
NetInterfaces -> Interfaces
getNetIOCountersAll -> getIOCountersAll
NetIOCounters -> IOCounters
NetIOCountersByFile -> IOCountersByFile
NetProtoCounters -> ProtoCounters
NetFilterCounters -> FilterCounters
NetConnections -> Connections
NetConnectionsPid -> ConnectionsPid
Uid -> UID
Id -> ID
convertCpuTimes -> convertCPUTimes
EOF`

IFS=$'\n'
for T in $TARGETS
do
  echo $T
  gofmt -w -r "$T" ./*.go
done


