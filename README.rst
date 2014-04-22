gopsutil: psutil for golang
==============================

This is a port of psutil(http://pythonhosted.org/psutil/). This
challenges port of all

Available archtectures
------------------------------------

- FreeBSD/amd64
- Linux
- Windows

(I do not have a darwin machine)

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

