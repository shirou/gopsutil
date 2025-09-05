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
    # Create the mount folder if it doesn't exist
    if (-not (Test-Path -Path $MountFolder)) {
        New-Item -Path $MountFolder -ItemType Directory | Out-Null
    }
    # Get the volume object
    if ($DriveLetter) {
        $vol = Get-Volume -DriveLetter $DriveLetter -ErrorAction SilentlyContinue
    } else {
        # Find the most recently created volume without a drive letter
        $vol = Get-Volume | Where-Object { $_.DriveLetter -eq $null } | Sort-Object -Property Size -Descending | Select-Object -First 1
    }
    if ($vol) {
            # Find the disk associated with the VHD
            $disk = Get-Disk | Where-Object { $_.Location -like "*${VhdPath}*" }
            if ($disk) {
                $part = Get-Partition -DiskNumber $disk.Number | Where-Object { $_.Type -eq 'IFS' } | Select-Object -First 1
                if ($part) {
                    Add-PartitionAccessPath -DiskNumber $disk.Number -PartitionNumber $part.PartitionNumber -AccessPath $MountFolder
                }
            }
    }
}

# Cleanup
Remove-Item $diskpartScriptPath -Force