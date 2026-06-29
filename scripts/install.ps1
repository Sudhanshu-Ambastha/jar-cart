param([string]$Version = "")
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrEmpty($Version)) {
    try {
        $info = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest"
        $Version = $info.tag_name
    } catch { $Version = "v0.0.1" }
}

$arch = "x86_64" # Assume x64 for Windows
$url = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version/jar-cart-$arch-windows.zip"
$zipPath = Join-Path ([System.IO.Path]::GetTempPath()) "jar-cart.zip"

Write-Host "⚡ Downloading $Version..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $zipPath

$installDir = Join-Path $HOME ".jar-cart\bin"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir -Force }

Write-Host "📦 Unpacking..."
Expand-Archive -Path $zipPath -DestinationPath $installDir -Force
Remove-Item $zipPath

$oldPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($oldPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$oldPath;$installDir", "User")
    Write-Host "🚀 Added to PATH." -ForegroundColor Green
}

Write-Host "✨ Done! Restart terminal to finish." -ForegroundColor Green