package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/shirou/gopsutil/v3/internal/common"
)

var ErrDockerNotAvailable = errors.New("docker not available")
var ErrCgroupNotAvailable = errors.New("cgroup not available")

var invoke common.Invoker = common.Invoke{}

const nanoseconds = 1e9

type CgroupDockerStat struct {
	ContainerID string `json:"containerID"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Status      string `json:"status"`
	Running     bool   `json:"running"`
}

func (c CgroupDockerStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

// getCgroupFilePath constructs file path to get targeted stats file.
func getCgroupFilePath(containerID, base, target, file string) string {
	if len(base) == 0 {
		base = common.HostSys(fmt.Sprintf("fs/cgroup/%s/docker", target))
	}
	statfile := path.Join(base, containerID, file)

	if _, err := os.Stat(statfile); os.IsNotExist(err) {
		statfile = path.Join(
			common.HostSys(fmt.Sprintf("fs/cgroup/%s/system.slice", target)), "docker-"+containerID+".scope", file)
	}

	return statfile
}