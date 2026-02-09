# gopsutil: psutil for Go

[![Test](https://github.com/shirou/gopsutil/actions/workflows/test.yml/badge.svg)](https://github.com/shirou/gopsutil/actions/workflows/test.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/shirou/gopsutil/v4.svg)](https://pkg.go.dev/github.com/shirou/gopsutil/v4) [![Calendar Versioning](https://img.shields.io/badge/calver-vMAJOR.YY.MM-22bfda.svg)](https://calver.org/)

This is a port of psutil (https://github.com/giampaolo/psutil). The
challenge is porting all psutil functions on some architectures.

## migration

### v4 migration

There are some breaking changes. Please see [v4 release note](https://github.com/shirou/gopsutil/releases/tag/v4.24.5).

## Tag semantics

gopsutil tag policy is almost same as Semantic Versioning, but
automatically increases like [Ubuntu versioning](https://calver.org/).

For example, v4.24.04 means

- v4: major version
- 24: release year, 2024
- 04: release month

gopsutil aims to keep backwards compatibility until major version change.

Tagged at every end of month, but if there are only a few commits, it
can be skipped.

## Available Architectures

- FreeBSD i386/amd64/arm
- Linux i386/amd64/arm(raspberry pi)
- Windows i386/amd64/arm/arm64
- Darwin amd64/arm64
- OpenBSD i386/amd64/armv7/arm64/riscv64 (Thank you @mpfz0r!)
- Solaris amd64 (developed and tested on SmartOS/Illumos, Thank you @jen20!)

These have partial support:

- CPU on DragonFly BSD (#893, Thank you @gballet!)
- host on Linux RISC-V (#896, Thank you @tklauser!)

All works are implemented without cgo by porting C structs to Go structs.

## Usage

```go
package main

import (
    "fmt"

    "github.com/shirou/gopsutil/v4/mem"
)

func main() {
    v, _ := mem.VirtualMemory()

    // almost every return value is a struct
    fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

    // convert to JSON. String() is also implemented
    fmt.Println(v)
}
```

The output is below.

    Total: 3179569152, Free:284233728, UsedPercent:84.508194%
    {"total":3179569152,"available":492572672,"used":2895335424,"usedPercent":84.50819439828305, (snip...)}

You can set alternative locations for various system directories by using the following environment variables:

- /proc: `HOST_PROC`
- /sys: `HOST_SYS`
- /etc: `HOST_ETC`
- /var: `HOST_VAR`
- /run: `HOST_RUN`
- /dev: `HOST_DEV`
- /: `HOST_ROOT`
- /proc/N/mountinfo: `HOST_PROC_MOUNTINFO`

### Adding settings using `context` (from v3.23.6)

As of v3.23.6, it is now possible to pass a path location using `context`: import `"github.com/shirou/gopsutil/v3/common"` and pass a context with `common.EnvMap` set to `common.EnvKey`, and the location will be used within each function.

```
	ctx := context.WithValue(context.Background(), 
		common.EnvKey, common.EnvMap{common.HostProcEnvKey: "/myproc"},
	)
	v, err := mem.VirtualMemoryWithContext(ctx)
```

First priority is given to the value set in `context`, then the value from the environment variable, and finally the default location.

### Caching

As of v3.24.1, it is now possible to cached some values. These values default to false, not cached. 

Be very careful that enabling the cache may cause inconsistencies. For example, if you enable caching of boottime on Linux, be aware that unintended values may be returned if [the boottime is changed by NTP after booted](https://github.com/shirou/gopsutil/issues/1070#issuecomment-842512782).

- `host`
  - EnableBootTimeCache
- `process`
  - EnableBootTimeCache

### `Ex` struct (from v4.24.5)

gopsutil is designed to work across multiple platforms. However, there are differences in the information available on different platforms, such as memory information that exists on Linux but not on Windows.

As of v4.24.5, to access this platform-specific information, gopsutil provides functions named `Ex` within the package. Currently, these functions are available in the mem and sensor packages.

The Ex structs are specific to each platform. For example, on Linux, there is an `ExLinux` struct, which can be obtained using the `mem.NewExLinux()` function. On Windows, it's `mem.ExWindows()`. These Ex structs provide platform-specific information.

```
ex := NewExWindows()
v, err := ex.VirtualMemory()
if err != nil {
    panic(err)
}

fmt.Println(v.VirtualAvail)
fmt.Println(v.VirtualTotal)

// Output:
// 140731958648832
// 140737488224256
```

gopsutil aims to minimize platform differences by offering common functions. However, there are many requests to obtain unique information for each platform. The Ex structs are designed to meet those requests. Additional functionalities might be added in the future. The use of these structures makes it clear that the information they provide is specific to each platform, which is why they have been designed in this way.

## Documentation

See https://pkg.go.dev/github.com/shirou/gopsutil/v4 or https://godocs.io/github.com/shirou/gopsutil/v4

## Requirements

- go1.18 or above is required.

## More Info

Several methods have been added which are not present in psutil, but
will provide useful information.

- host/HostInfo() (linux)
  - Hostname
  - Uptime
  - Procs
  - OS (ex: "linux")
  - Platform (ex: "ubuntu", "arch")
  - PlatformFamily (ex: "debian")
  - PlatformVersion (ex: "Ubuntu 13.10")
  - VirtualizationSystem (ex: "LXC")
  - VirtualizationRole (ex: "guest"/"host")
- IOCounters
  - Label (linux only) The registered [device mapper
    name](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-block-dm)
- cpu/CPUInfo() (linux, freebsd)
  - CPU (ex: 0, 1, ...)
  - VendorID (ex: "GenuineIntel")
  - Family
  - Model
  - Stepping
  - PhysicalID
  - CoreID
  - Cores (ex: 2)
  - ModelName (ex: "Intel(R) Core(TM) i7-2640M CPU @ 2.80GHz")
  - Mhz
  - CacheSize
  - Flags (ex: "fpu vme de pse tsc msr pae mce cx8 ...")
  - Microcode
- load/Avg() (linux, freebsd, solaris)
  - Load1
  - Load5
  - Load15
- docker/GetDockerIDList() (linux only)
  - container id list ([]string)
- docker/CgroupCPU() (linux only)
  - user
  - system
- docker/CgroupMem() (linux only)
  - various status
- net_protocols (linux only)
  - system wide stats on network protocols (i.e IP, TCP, UDP, etc.)
  - sourced from /proc/net/snmp
- iptables nf_conntrack (linux only)
  - system wide stats on netfilter conntrack module
  - sourced from /proc/sys/net/netfilter/nf_conntrack_count

Some code is ported from Ohai. Many thanks.

## Current Status

- x: works
- b: almost works, but something is broken

|name                  |Linux  |FreeBSD  |OpenBSD  |macOS   |Windows  |Solaris  |Plan 9   |AIX      |
|----------------------|-------|---------|---------|--------|---------|---------|---------|---------|
|cpu\_times            |x      |x        |x        |x       |x        |         |b        |x        |
|cpu\_count            |x      |x        |x        |x       |x        |         |x        |x        |
|cpu\_percent          |x      |x        |x        |x       |x        |         |         |x        |
|cpu\_times\_percent   |x      |x        |x        |x       |x        |         |         |x        |
|virtual\_memory       |x      |x        |x        |x       |x        |b        |x        |x        |
|swap\_memory          |x      |x        |x        |x       |         |         |x        |X        |
|disk\_partitions      |x      |x        |x        |x       |x        |         |         |x        |
|disk\_io\_counters    |x      |x        |x        |x       |x        |         |         |         |
|disk\_usage           |x      |x        |x        |x       |x        |         |         |x        |
|net\_io\_counters     |x      |x        |x        |b       |x        |x        |         |         |
|boot\_time            |x      |x        |x        |x       |x        |         |         |X        |
|users                 |x      |x        |x        |x       |x        |         |         |x        |
|pids                  |x      |x        |x        |x       |x        |         |         |         |
|pid\_exists           |x      |x        |x        |x       |x        |         |         |         |
|net\_connections      |x      |x        |x        |x       |         |         |         |x        |
|net\_protocols        |x      |         |         |        |         |         |         |x        |
|net\_if\_addrs        |       |         |         |        |         |         |         |x        |
|net\_if\_stats        |       |         |         |        |         |         |         |x        |
|netfilter\_conntrack  |x      |         |         |        |         |         |         |         |
|sensors_temperature   |x      |         |         |x       |x        |x        |         |         |


### Process class

|name                |Linux  |FreeBSD  |OpenBSD  |macOS  |Windows  |
|--------------------|-------|---------|---------|-------|---------|
|pid                 |x      |x        |x        |x      |x        |
|ppid                |x      |x        |x        |x      |x        |
|name                |x      |x        |x        |x      |x        |
|cmdline             |x      |x        |x        |x      |x        |
|create\_time        |x      |         |         |x      |x        |
|status              |x      |x        |x        |x      |         |
|cwd                 |x      |x        |x        |x      |x        |
|exe                 |x      |x        |x        |       |x        |
|uids                |x      |x        |x        |x      |         |
|gids                |x      |x        |x        |x      |         |
|terminal            |x      |x        |x        |       |         |
|io\_counters        |x      |x        |x        |       |x        |
|nice                |x      |x        |x        |x      |x        |
|num\_fds            |x      |         |         |x      |x        |
|num\_ctx\_switches  |x      |         |         |       |         |
|num\_threads        |x      |x        |x        |x      |x        |
|cpu\_times          |x      |         |         |       |x        |
|memory\_info        |x      |x        |x        |x      |x        |
|memory\_maps        |x      |         |         |       |         |
|open\_files         |x      |         |         |       |x        |
|send\_signal        |x      |x        |x        |x      |         |
|suspend             |x      |x        |x        |x      |x        |
|resume              |x      |x        |x        |x      |x        |
|terminate           |x      |x        |x        |x      |x        |
|kill                |x      |x        |x        |x      |         |
|username            |x      |x        |x        |x      |x        |
|ionice              |       |         |         |       |         |
|rlimit              |x      |         |         |       |         |
|num\_handlers       |       |         |         |       |         |
|threads             |x      |         |         |       |         |
|cpu\_percent        |x      |         |x        |x      |x        |
|cpu\_affinity       |       |         |         |       |         |
|memory\_percent     |x      |         |         |       |x        |
|parent              |x      |         |x        |x      |x        |
|children            |x      |x        |x        |x      |x        |
|connections         |x      |         |x        |x      |         |
|is\_running         |       |         |         |       |         |
|page\_faults        |x      |         |         |       |x        |

### gopsutil Original Metrics

|item                    |Linux  |FreeBSD  |OpenBSD  |macOS   |Windows |Solaris  |AIX      |
|------------------------|-------|---------|---------|--------|--------|---------|---------|
|**HostInfo**            |       |         |         |        |        |         |         |
|hostname                |x      |x        |x        |x       |x       |x        |X        |
|uptime                  |x      |x        |x        |x       |        |x        |x        |
|process                 |x      |x        |x        |        |        |x        |         |
|os                      |x      |x        |x        |x       |x       |x        |x        |
|platform                |x      |x        |x        |x       |        |x        |x        |
|platformfamily          |x      |x        |x        |x       |        |x        |x        |
|virtualization          |x      |         |         |        |        |         |         |
|**CPU**                 |       |         |         |        |        |         |         |
|VendorID                |x      |x        |x        |x       |x       |x        |x        |
|Family                  |x      |x        |x        |x       |x       |x        |x        |
|Model                   |x      |x        |x        |x       |x       |x        |x        |
|Stepping                |x      |x        |x        |x       |x       |x        |         |
|PhysicalID              |x      |         |         |        |        |x        |         |
|CoreID                  |x      |         |         |        |        |x        |         |
|Cores                   |x      |         |         |x       |x       |x        |x        |
|ModelName               |x      |x        |x        |x       |x       |x        |x        |
|Microcode               |x      |         |         |        |        |x        |         |
|**LoadAvg**             |       |         |         |        |        |         |         |
|Load1                   |x      |x        |x        |x       |x       |x        |x        |
|Load5                   |x      |x        |x        |x       |x       |x        |x        |
|Load15                  |x      |x        |x        |x       |x       |x        |x        |
|**Docker GetDockerID**  |       |         |         |        |        |         |         |
|container id            |x      |no       |no       |no      |no      |no       |no       |
|**Docker CgroupsCPU**   |       |         |         |        |        |         |         |
|user                    |x      |no       |no       |no      |no      |no       |no       |
|system                  |x      |no       |no       |no      |no      |no       |no       |
|**Docker CgroupsMem**   |       |         |         |        |        |         |         |
|various                 |x      |no       |no       |no      |no      |no       |no       |

- future work
  - process_iter
  - wait_procs
  - Process class
    - as_dict
    - wait
  - AIX processes

## License

New BSD License (same as psutil)

## Related Works

I have been influenced by the following great works:

- psutil: https://github.com/giampaolo/psutil
- dstat: https://github.com/dagwieers/dstat
- gosigar: https://github.com/cloudfoundry/gosigar/
- goprocinfo: https://github.com/c9s/goprocinfo
- go-ps: https://github.com/mitchellh/go-ps
- ohai: https://github.com/opscode/ohai/
- bosun:
  https://github.com/bosun-monitor/bosun/tree/master/cmd/scollector/collectors
- mackerel:
  https://github.com/mackerelio/mackerel-agent/tree/master/metrics

## How to Contribute

1.  Fork it
2.  Create your feature branch (git checkout -b my-new-feature)
3.  Commit your changes (git commit -am 'Add some feature')
4.  Push to the branch (git push origin my-new-feature)
5.  Create new Pull Request

English is not my native language, so PRs correcting grammar or spelling are welcome and appreciated.
