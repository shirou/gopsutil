param(
    [Parameter(Mandatory=$true)]
    [string]$VhdPath,
    [Parameter(Mandatory=$false)]
    [string]$DriveLetter,
    [Parameter(Mandatory=$false)]
    [string]$MountFolder
)

# Create diskpart script
$diskpartScript = @"
create vdisk file="$VhdPath" maximum=10 type=expandable
select vdisk file="$VhdPath"
attach vdisk
create partition primary
format fs=ntfs quick`r`n
"@

if ($DriveLetter) {
    $diskpartScript += "assign letter=$DriveLetter`r`n"
}

$diskpartScriptPath = "$env:TEMP\diskpart_script.txt"
$diskpartScript | Set-Content -Path $diskpartScriptPath

# Run diskpart
diskpart /s $diskpartScriptPath

# Mount to folder if specified
if ($MountFolder) {
    # Get the volume object
    $vol = Get-WmiObject -Query "SELECT * FROM Win32_Volume WHERE DriveLetter = '${DriveLetter}:'"
    if ($vol) {
        $vol.AddMountPoint($MountFolder)
    }
}

# Cleanup
Remove-Item $diskpartScriptPath -Force