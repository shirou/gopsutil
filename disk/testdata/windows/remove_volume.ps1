param(
    [Parameter(Mandatory=$true)]
    [string]$VhdPath,
    [Parameter(Mandatory=$false)]
    [string]$MountFolder
)

# Detach the VHD
$diskpartScript = @"
select vdisk file="$VhdPath"
detach vdisk
"@
$diskpartScriptPath = "$env:TEMP\diskpart_remove_script.txt"
$diskpartScript | Set-Content -Path $diskpartScriptPath
diskpart /s $diskpartScriptPath
Remove-Item $diskpartScriptPath -Force

# Delete the VHD file
if (Test-Path $VhdPath) {
    Remove-Item $VhdPath -Force
}

# Remove mount folder if specified
if ($MountFolder -and (Test-Path $MountFolder)) {
    Remove-Item $MountFolder -Recurse -Force
}