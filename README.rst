gopsutil: psutil for golang
==============================

.. image:: https://drone.io/github.com/shirou/gopsutil/status.png
        :target: https://drone.io/github.com/shirou/gopsutil
		
This is a port of psutil(http://pythonhosted.org/psutil/). This
challenges porting all psutil functions on some architectures.

Available archtectures
------------------------------------

- FreeBSD/amd64
- Linux/amd64
- Windows/amd64

(I do not have a darwin machine)


All works are implemented without cgo by porting c struct to golang struct.


usage
---------

::

  import (
      "github.com/shirou/gopsutil"
      "fmt"
      "encoding/json"
  )

  func main(){
      v, _ := gopsutil.Virtual_memory()

      // return value is struct
      fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

      // convert to JSON
      d, _ := json.Marshal(v)
      fmt.Printf("%s\n", d)
  }

The output is below.

::

  Total: 3179569152, Free:284233728, UsedPercent:84.508194%
  {"total":3179569152,"available":492572672,"used":2895335424,"usedPercent":84.50819439828305, (snip)}


Document
----------

see http://godoc.org/github.com/shirou/gopsutil


Current Status
------------------

- done

  - cpu_times (linux, freebsd)
  - cpu_count (linux, freebsd, windows)
  - virtual_memory (linux, windows)
  - swap_memory (linux)
  - disk_partitions (freebsd, windows)
  - disk_usage (linux, freebsd, windows)
  - boot_time (linux, freebsd, windows(but little broken))
  - users (linux, freebsd)
  - pids (linux, freebsd)
  - pid_exists (linux, freebsd)

- not yet

  - cpu_percent
  - cpu_times_percent
  - disk_io_counters
  - net_io_counters
  - net_connections
  - Process class

- future work

  - process_iter
  - wait_procs



License
------------

New BSD License (same as psutil)


Related works
-----------------------

So many thanks!

- psutil: http://pythonhosted.org/psutil/
- dstat: https://github.com/dagwieers/dstat
- gosiger: https://github.com/cloudfoundry/gosigar/
- goprocinfo: https://github.com/c9s/goprocinfo
- go-ps: https://github.com/mitchellh/go-ps

