// +build linux

package process

import (
	"context"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func Test_fillFromStatusWithContext(t *testing.T) {
	pids, err := ioutil.ReadDir("testdata/linux/")
	if err != nil {
		t.Error(err)
	}
	f := common.MockEnv("HOST_PROC", "testdata/linux")
	defer f()
	for _, pid := range pids {
		pid, _ := strconv.ParseInt(pid.Name(), 0, 32)
		p, _ := NewProcess(int32(pid))

		if err := p.fillFromStatusWithContext(context.Background()); err != nil {
			t.Error(err)
		}
	}
}
