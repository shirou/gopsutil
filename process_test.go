package gopsutil

import (
	"testing"
)

func Test_Pids(t *testing.T) {
	ret, err := Pids()
	if err != nil {
		t.Errorf("error %v", err)
	}
	if len(ret) == 0 {
		t.Errorf("could not get pids %v", ret)
	}
}

func Test_Pid_exists(t *testing.T) {
	ret, err := Pid_exists(1)
	if err != nil {
		t.Errorf("error %v", err)
	}

	if ret == false {
		t.Errorf("could not get init process %v", ret)
	}

}
