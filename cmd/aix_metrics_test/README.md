# AIX Metrics Test Program

This is a comprehensive wrapper test program that exercises all gopsutil metrics supported on AIX and outputs them to a text file.

## Building

```bash
# Ensure you have Go installed on AIX
cd /path/to/gopsutil/cmd/aix_metrics_test
go build -o aix_metrics_test main.go
```

## Running

```bash
./aix_metrics_test
```

The program will create a file named `gopsutil_metrics.txt` in the current directory containing all supported metrics.

## Output

The program outputs the following categories of metrics:

- **Host Information**: Hostname, uptime, OS, platform, kernel details, virtualization
- **CPU Information**: CPU counts, info, times, and utilization percentages
- **Memory Information**: Virtual and swap memory statistics
- **Load Average**: System load averages and process counts
- **Disk Information**: Partitions and disk usage
- **Network Information**: Network interfaces and I/O counters
- **Process Information**: Current process details and all processes on the system

## Notes

- This program is designed specifically for AIX systems (build flag: `//go:build aix`)
- Some metrics may not be fully implemented on AIX and will return "ErrNotImplementedError"
- The output includes both console output and file output for comprehensive logging
- CPU metrics are gathered with a 1-second sampling window

## Supported Metrics

The program attempts to retrieve all available metrics from:
- `cpu` package
- `disk` package
- `host` package
- `load` package
- `mem` package
- `net` package
- `process` package

Refer to the gopsutil documentation for detailed descriptions of each metric.
