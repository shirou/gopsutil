package disk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogicalDrives(t *testing.T) {
	drives, err := getLogicalDrives()
	require.NoError(t, err)
	assert.NotEmpty(t, drives)
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
	assert.Equal(t, "NTFS", part.Fstype) // NTFS should be the only allowed fs on C: drive since windows Vista, maybe in future could be also reFS
	assert.Contains(t, part.Opts, rw)    // C: must have atleast rw option
}

func TestProcessLogicalDrives(t *testing.T) {
	drives := []string{`C:\`}
	partitionStats := []PartitionStat{}
	processedPaths := map[string]struct{}{}
	warnings := Warnings{}

	parts := processLogicalDrives(drives, processedPaths, partitionStats, warnings)
	assert.Len(t, parts, 1)
	assert.Equal(t, "C:", parts[0].Mountpoint)
	assert.Equal(t, "C:", parts[0].Device)
	assert.Equal(t, "NTFS", parts[0].Fstype)
	assert.Contains(t, parts[0].Opts, rw)
}
