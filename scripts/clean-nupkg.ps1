param (
    [string]$TargetDir = "jc"
)

Write-Host "Checking for legacy .nupkg files in '$TargetDir'..."
$oldPackages = Get-ChildItem -Path $TargetDir -Filter "*.nupkg"

if ($oldPackages) {
    Write-Host "Found old .nupkg files. Cleaning up to prevent conflicts..."
    Remove-Item -Path "$TargetDir/*.nupkg" -Force
    Write-Host "Old packages removed successfully."
} else {
    Write-Host "No legacy .nupkg files found. Skipping cleanup."
}