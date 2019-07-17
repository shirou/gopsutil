// +build linux

package process

import (
	"testing"
)

func Test_Process_MemoryInfoSmaps(t *testing.T) {
	p := testGetProcess()
	v, err := p.MemoryInfoSmaps()
	if err != nil {
		t.Errorf("geting extended memory info error %v", err)
	}
	empty := MemoryInfoSmapsStat{}
	if v == nil || *v == empty {
		t.Errorf("could not get extended memory info %v", v)
	}
}
