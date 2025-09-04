package disk

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const (
	createVolumeScript = `.\testdata\windows\create_volume.ps1`
	removeVolumeScript = `.\testdata\windows\remove_volume.ps1`
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

func TestGetPartStatFromVolumeName(t *testing.T) {
	driveLetter := "Y"
	vhdFile := `C:\testdisk.vhd`
	mountFolder := `C:\mountpoint`
	removeMountedVolume(t, vhdFile, mountFolder)

	mountVolume(t, driveLetter, vhdFile, mountFolder)
	warnings := Warnings{}
	processedPaths := make(map[string]struct{})
	partitionsStats := make([]PartitionStat, 0)
	drivePath := driveLetter + `:\`
	volNameBuffer := windows.StringToUTF16(drivePath)
	getPartStatFromVolumeName(context.Background(), volNameBuffer, &warnings, processedPaths, partitionsStats)
	assert.Empty(t, warnings.List)
}

func mountVolume(t *testing.T, driveLetter string, vhdFile string, mountFolder string) {
	mountVolumeCmd := exec.Command("powershell.exe", "-File", createVolumeScript, "-DriveLetter", driveLetter, "-VhdPath", vhdFile, "-MountFolder", mountFolder)
	out, createVolumeErr := mountVolumeCmd.Output()
	require.NoError(t, createVolumeErr, out)
}

func removeMountedVolume(t *testing.T, vhdFile string, mountFolder string) {
	if _, statErr := os.Stat(vhdFile); statErr == nil {
		unmountVolumeCmd := exec.Command("powershell.exe", "-File", removeVolumeScript, "-VhdPath", vhdFile, "-MountFolder", mountFolder)
		out, unmountVolumeErr := unmountVolumeCmd.CombinedOutput()
		require.NoError(t, unmountVolumeErr, out)
	}
}
