package cgroup

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/shirou/gopsutil/v3/internal/common"
)


var initCGroup = &sync.Once{}

func getBaseCGroup() string {
	initCGroup.Do(func(){
		if os.Getenv("HOST_CGROUP") == "" {
			os.Setenv("HOST_CGROUP", "fs/cgroup")
		}
	})

	return os.Getenv("HOST_CGROUP")
}

// getCgroupFilePath constructs file path to get targeted stats file.
func getCgroupFilePath(base, target, file string) string {
	if len(base) == 0 {
		base = common.HostSys(path.Join(getBaseCGroup(), target))
	}
	statfile := path.Join(base, file)

	if _, err := os.Stat(statfile); os.IsNotExist(err) {
		statfile = path.Join(
			common.HostSys(fmt.Sprintf("fs/cgroup/%s/system.slice", target)), file)
	}

	return statfile
}

// getCgroupMemFile reads a cgroup file and return the contents as uint64.
func getCgroupMemFile(base, file string) (uint64, error) {
	statfile := getCgroupFilePath(base, "memory", file)
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return 0, err
	}
	if len(lines) != 1 {
		return 0, fmt.Errorf("wrong format file: %s", statfile)
	}
	return strconv.ParseUint(lines[0], 10, 64)
}
