// +build linux

package process

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func Test_fillFromStatusWithContext(t *testing.T) {
	pids, err := ioutil.ReadDir("testdata/linux/")
	if err != nil {
		t.Error(err)
	}
	original := os.Getenv("HOST_PROC")
	os.Setenv("HOST_PROC", "testdata/linux")
	defer os.Setenv("HOST_PROC", original)

	for _, pid := range pids {
		pid, _ := strconv.ParseInt(pid.Name(), 0, 32)
		p, _ := NewProcess(int32(pid))

		if err := p.fillFromStatusWithContext(context.Background()); err != nil {
			t.Error(err)
		}
	}
}
