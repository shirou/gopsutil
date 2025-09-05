package disk

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	mountFolder := `C:\mountpoint\`
	longMountFolder := `C:\this\is\a\very\long\mountpoint\to\test\the\maximum\path\length\allowed\by\windows\operating\system\for\mounted\volumes\as\folders\in\the\file\system\structure\of\the\os\itself\and\see\if\there\are\any\issues\with\paths\longer\than\the\traditional\260\character\limit\which\was\present\in\older\versions\of\windows\operating\system\and\is\still\a\common\issue\for\many\applications\ok\this\is\just\a\very\long\text\to\test\if\we\are\calling\the\api\correctly\it\is\not\exptected\that\you\read\it`
	testMountedVolumesAsFolder(t, driveLetter, "")
	testMountedVolumesAsFolder(t, driveLetter, mountFolder)
	testMountedVolumesAsFolder(t, "", mountFolder)
	testMountedVolumesAsFolder(t, driveLetter, longMountFolder)
	testMountedVolumesAsFolder(t, "", longMountFolder)
}

func testMountedVolumesAsFolder(t *testing.T, driveLetter string, mountFolder string) {
	fsType := "NTFS"
	vhdFile := `C:\testdisk.vhd`
	opts := []string{"rw", "compress"}
	letterMountPoint := driveLetter + `:\`
	defer removeMountedVolume(t, vhdFile, mountFolder)

	mountVolume(t, driveLetter, vhdFile, mountFolder)
	warnings := Warnings{}
	processedPaths := make(map[string]struct{})
	partitionsStats := make([]PartitionStat, 0)
	partitionStats := processVolumesMountedAsFolders(context.Background(), partitionsStats, processedPaths, &warnings)
	assert.Empty(t, warnings.List)
	assert.Greater(t, len(partitionStats), 1)
	if mountFolder != "" {
		assert.Contains(t, partitionStats, PartitionStat{Mountpoint: mountFolder, Device: mountFolder, Fstype: fsType, Opts: opts})
	}
	if driveLetter != "" {
		assert.Contains(t, partitionStats, PartitionStat{Mountpoint: letterMountPoint, Device: letterMountPoint, Fstype: fsType, Opts: opts})
	}
}

func mountVolume(t *testing.T, driveLetter string, vhdFile string, mountFolder string) {
	args := []string{"-File", createVolumeScript, "-VhdPath", vhdFile}
	if driveLetter != "" {
		args = append(args, "-DriveLetter", driveLetter)
	}
	if mountFolder != "" {
		args = append(args, "-MountFolder", mountFolder)
	}
	mountVolumeCmd := exec.Command("powershell.exe", args...)
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
