package disk

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogicalDrives(t *testing.T) {
	// Create a virtual drive using subst to ensure multiple drives exist
	tempDir := t.TempDir()

	// Find an unused drive letter (Y: to G:, reverse order)
	var driveLetter string
	for c := 'Y'; c >= 'G'; c-- {
		testDrive := string(c) + ":"
		_, err := os.Stat(testDrive + "\\")
		if os.IsNotExist(err) {
			driveLetter = testDrive
			break
		}
	}
	if driveLetter == "" {
		t.Skip("No available drive letter for subst")
	}
	ctx := context.Background()
	// Create virtual drive by using subst command
	cmd := exec.CommandContext(ctx, "subst", driveLetter, tempDir)
	if err := cmd.Run(); err != nil {
		t.Skipf("subst command failed: %v", err)
	}
	t.Cleanup(func() {
		exec.CommandContext(ctx, "subst", driveLetter, "/d").Run()
	})

	drives, err := getLogicalDrives(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, drives)
	for _, d := range drives {
		assert.NotEmpty(t, d)
	}
	t.Log("Logical Drives:", drives)
	assert.Contains(t, drives, `C:\`)
	assert.Contains(t, drives, driveLetter+`\`)
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

	parts := processLogicalDrives(context.Background(), drives, processedPaths, partitionStats, warnings)
	assert.Len(t, parts, 1)
	assert.Equal(t, "C:", parts[0].Mountpoint)
	assert.Equal(t, "C:", parts[0].Device)
	assert.Equal(t, "NTFS", parts[0].Fstype)
	assert.Contains(t, parts[0].Opts, rw)
}
