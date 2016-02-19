// +build darwin

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var vmStatOut = `
Mach Virtual Memory Statistics: (page size of 4096 bytes)
Pages free:                              105885.
Pages active:                            725641.
Pages inactive:                          449242.
Pages speculative:                         6155.
Pages throttled:                              0.
Pages wired down:                        560835.
Pages purgeable:                         128967.
"Translation faults":                 622528839.
Pages copy-on-write:                   17697839.
Pages zero filled:                    311034413.
Pages reactivated:                      4705104.
Pages purged:                           5605610.
File-backed pages:                       349192.
Anonymous pages:                         831846.
Pages stored in compressor:              876507.
Pages occupied by compressor:            249167.
Decompressions:                         4555025.
Compressions:                           7524729.
Pageins:                               40532443.
Pageouts:                                126496.
Swapins:                                2988073.
Swapouts:                               3283599.
`

func TestParseVmStat(t *testing.T) {
	ret := &VirtualMemoryStat{}
	err := parseVmStat(vmStatOut, 4096, ret)

	if err != nil {
		t.Errorf("Expected no error, got %s\n", err.Error())
	}

	if ret.Free != uint64(105885*4096) {
		t.Errorf("Free pages, actual: %d, expected: %d", ret.Free,
			105885*4096)
	}

	if ret.Inactive != uint64(449242*4096) {
		t.Errorf("Inactive pages, actual: %d, expected: %d", ret.Inactive,
			449242*4096)
	}

	if ret.Active != uint64(725641*4096) {
		t.Errorf("Active pages, actual: %d, expected: %d", ret.Active,
			725641*4096)
	}

	if ret.Wired != uint64(560835*4096) {
		t.Errorf("Wired pages, actual: %d, expected: %d", ret.Wired,
			560835*4096)
	}

	if ret.Cached != uint64(128967*4096+449242.*4096) {
		t.Errorf("Cached pages, actual: %d, expected: %f", ret.Cached,
			128967*4096+449242.*4096)
	}
}

func TestVirtualMemoryDarwin(t *testing.T) {
	v, err := VirtualMemory()
	assert.Nil(t, err)

	assert.True(t, v.Total > 0)

	assert.True(t, v.Available > 0)
	assert.True(t, v.Available < v.Total)
	assert.Equal(t, v.Available, v.Total-v.Wired-v.Active, "%v", v)

	assert.True(t, v.Used > 0)
	assert.True(t, v.Used < v.Total)

	assert.True(t, v.UsedPercent > 0)
	assert.True(t, v.UsedPercent < 100)

	assert.True(t, v.Free > 0)
	assert.True(t, v.Free < v.Available)

	assert.True(t, v.Active > 0)
	assert.True(t, v.Active < v.Total)

	assert.True(t, v.Inactive > 0)
	assert.True(t, v.Inactive < v.Total)

	assert.True(t, v.Wired > 0)
	assert.True(t, v.Wired < v.Total)
}
