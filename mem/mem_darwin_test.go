// +build darwin

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
