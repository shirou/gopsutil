# v2 to v3 changes


- create v3 directory
- Remove process.NetIOCounters (#429)
- rename memoryLimitInBbytes JSON key in docker (#464)
- fix cgroup filename (#464)
- RLimit is now uint64 (#364)
- mem.VirtualMemoryStat JSON fields capitalization (#545)
  - various JSON field name and some of Variable name have been changed. see v3migration.sh
- Become private various kind of platform dependent values/constants such as process.GetWin32Proc.

### not yet

- Determine if process is running in the foreground (#596)
