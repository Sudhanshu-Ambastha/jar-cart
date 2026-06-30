param([string]$Version = "")
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrEmpty($Version)) {
    Write-Host "🔍 Fetching latest version from GitHub..." -ForegroundColor Cyan
    try {
        $info = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest"
        $Version = $info.tag_name
    } catch {
        Write-Error "❌ Failed to fetch latest version. Please provide a version manually (e.g., -Version v0.1.0)"
        exit 1
    }
}

$arch = "x86_64" 
$url = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version/jar-cart-$arch-windows.zip"
$zipPath = Join-Path ([System.IO.Path]::GetTempPath()) "jar-cart.zip"

Write-Host "⚡ Downloading $Version..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $url -OutFile $zipPath
} catch {
    Write-Error "❌ Download failed. Version $Version might not exist or network is down."
    exit 1
}

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

Write-Host "✨ Done! Restart your terminal to use jar-cart." -ForegroundColor Green