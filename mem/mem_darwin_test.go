// SPDX-License-Identifier: BSD-3-Clause
//go:build darwin

package mem

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualMemoryDarwin(t *testing.T) {
	v, err := VirtualMemory()
	require.NoError(t, err)

	outBytes, err := invoke.Command("/usr/sbin/sysctl", "hw.memsize")
	require.NoError(t, err)
	outString := string(outBytes)
	outString = strings.TrimSpace(outString)
	outParts := strings.Split(outString, " ")
	actualTotal, err := strconv.ParseInt(outParts[1], 10, 64)
	require.NoError(t, err)
	assert.Equal(t, uint64(actualTotal), v.Total)

	assert.Positive(t, v.Available)
	assert.Equal(t, v.Available, v.Free+v.Inactive, "%v", v)

	assert.Positive(t, v.Used)
	assert.Less(t, v.Used, v.Total)

	assert.Positive(t, v.UsedPercent)
	assert.Less(t, v.UsedPercent, 100.0)

	assert.Positive(t, v.Free)
	assert.Less(t, v.Free, v.Available)

	assert.Positive(t, v.Active)
	assert.Less(t, v.Active, v.Total)

	assert.Positive(t, v.Inactive)
	assert.Less(t, v.Inactive, v.Total)

	assert.Positive(t, v.Wired)
	assert.Less(t, v.Wired, v.Total)
}
