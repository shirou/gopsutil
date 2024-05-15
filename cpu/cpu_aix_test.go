package cpu

import (
	"testing"
)

func TestTimes(t *testing.T) {
	_, err := Times(true)
	if err != nil {
		t.Error("Times(true) failed")
	}
	_, err = Times(false)
	if err != nil {
		t.Error("Times(false) failed")
	}
}

func TestInfo(t *testing.T) {
	{
		l, err := Times(true)
		if err != nil || len(l) == 0 {
			t.Error("Times(true) failed")
		}
		t.Logf("Times(true): %#v", l)
	}
	{
		l, err := Times(false)
		if err != nil || len(l) == 0 {
			t.Error("Times(false) failed")
		}
		t.Logf("Times(false): %#v", l)
	}
}
