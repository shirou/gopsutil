package mem

import (
	"testing"
)

func TestVirtualMemoryEx(t *testing.T) {
	v, err := VirtualMemoryEx()
	if err != nil {
		t.Error(err)
	}

	t.Log(v)
}
