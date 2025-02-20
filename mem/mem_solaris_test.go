// SPDX-License-Identifier: BSD-3-Clause
//go:build solaris

package mem

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validFile = `swapfile                  dev  swaplo blocks   free
/dev/zvol/dsk/rpool/swap 256,1      16 1058800 1058800
/dev/dsk/c0t0d0s1   136,1      16 1638608 1600528`

const invalidFile = `swapfile                  dev  swaplo INVALID   free
/dev/zvol/dsk/rpool/swap 256,1      16 1058800 1058800
/dev/dsk/c0t0d0s1   136,1      16 1638608 1600528`

func TestParseSwapsCommandOutput_Valid(t *testing.T) {
	stats, err := parseSwapsCommandOutput(validFile)
	require.NoError(t, err)

	assert.Equal(t, SwapDevice{
		Name:      "/dev/zvol/dsk/rpool/swap",
		UsedBytes: 0,
		FreeBytes: 1058800 * 512,
	}, *stats[0])

	assert.Equal(t, SwapDevice{
		Name:      "/dev/dsk/c0t0d0s1",
		UsedBytes: 38080 * 512,
		FreeBytes: 1600528 * 512,
	}, *stats[1])
}

func TestParseSwapsCommandOutput_Invalid(t *testing.T) {
	_, err := parseSwapsCommandOutput(invalidFile)
	assert.Error(t, err)
}

func TestParseSwapsCommandOutput_Empty(t *testing.T) {
	_, err := parseSwapsCommandOutput("")
	assert.Error(t, err)
}
