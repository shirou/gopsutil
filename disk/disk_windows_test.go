package disk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogicalDrives(t *testing.T) {
	ctx := context.Background()
	drives, err := getLogicaldrives(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(drives), 0)
	for _, d := range drives {
		assert.NotEmpty(t, d)
	}
	assert.Contains(t, drives, `C:\`)
}

func TestBuildPartitionStat(t *testing.T) {
	volumeC := `C:\`
	part, err := buildPartitionStat(volumeC)
	require.NoError(t, err)
	assert.Equal(t, volumeC, part.Mountpoint)
	assert.Equal(t, volumeC, part.Device)
	assert.Equal(t, "NTFS", part.Fstype) // NTFS should be the only allowed fs on C: drive
	assert.Contains(t, part.Opts, "rw")  // C: must have atleast rw option
}
