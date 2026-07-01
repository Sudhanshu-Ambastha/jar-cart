$cacheDir = "$env:USERPROFILE\.jar-cart\cache"
$jdksDir = "$env:USERPROFILE\.jar-cart\jdks"

Write-Host "=== Analyzing Jar-Cart Ecosystem ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "[Dependency Cache]" -ForegroundColor Green
if (!(Test-Path $cacheDir)) {
    Write-Host "Cache directory not found." -ForegroundColor Red
} else {
    $jarFiles = Get-ChildItem -Path $cacheDir -Filter "*.jar" -Recurse
    if ($jarFiles.Count -eq 0) {
        Write-Host "Cache is empty." -ForegroundColor Yellow
    } else {
        foreach ($jar in $jarFiles) {
            $group = Split-Path $jar.DirectoryName -Leaf
            Write-Host " Group: $group" -ForegroundColor Yellow
            Write-Host "  File: $($jar.Name)" -ForegroundColor White
        }
    }
}

Write-Host ""
Write-Host "[Available JDKs]" -ForegroundColor Green
if (!(Test-Path $jdksDir)) {
    Write-Host "JDK directory not found." -ForegroundColor Red
} else {
    $jdks = Get-ChildItem -Path $jdksDir -Directory
    if ($jdks.Count -eq 0) {
        Write-Host "No JDKs provisioned." -ForegroundColor Yellow
    } else {
        foreach ($jdk in $jdks) {
            Write-Host " Java Version: $($jdk.Name)" -ForegroundColor White -NoNewline
            Write-Host " (Path: $($jdk.FullName))" -ForegroundColor Gray
        }
    }
}