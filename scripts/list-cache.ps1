$cacheDir = "$env:USERPROFILE\.jar-cart\cache"

if (!(Test-Path $cacheDir)) {
    Write-Host "Cache directory not found at $cacheDir" -ForegroundColor Red
    exit
}

Write-Host "Analyzing Jar-Cart Cache..." -ForegroundColor Cyan
Write-Host "---------------------------------------------------"

$jarFiles = Get-ChildItem -Path $cacheDir -Filter "*.jar" -Recurse

if ($jarFiles.Count -eq 0) {
    Write-Host "Cache is empty." -ForegroundColor Yellow
}

foreach ($jar in $jarFiles) {
    $group = Split-Path $jar.DirectoryName -Leaf
    
    Write-Host "Group: $group" -ForegroundColor Yellow
    Write-Host "  File: $($jar.Name)" -ForegroundColor White
    Write-Host "  Path: $($jar.FullName)" -ForegroundColor Gray
    Write-Host ""
}