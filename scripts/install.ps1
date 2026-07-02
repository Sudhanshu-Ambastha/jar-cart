param([string]$Version = "")
$ErrorActionPreference = "Stop"

$arch = "x86_64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $arch = "aarch64" }

if ([string]::IsNullOrEmpty($Version)) {
    Write-Host "🔍 Fetching latest version from GitHub..." -ForegroundColor Cyan
    try {
        $info = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" -Headers @{"Accept"="application/vnd.github+json"}
        $Version = $info.tag_name
    } catch {
        Write-Error "❌ Failed to fetch latest version."
        exit 1
    }
}

$baseUrl = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version"
$zipName = "jar-cart-$arch-windows.zip"
$zipUrl = "$baseUrl/$zipName"
$hashUrl = "$zipUrl.sha256" 

$tempDir = [System.IO.Path]::GetTempPath()
$zipPath = Join-Path $tempDir $zipName
$hashPath = Join-Path $tempDir "$zipName.sha256"

Write-Host "⚡ Downloading $Version ($arch)..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath -MaximumRedirection 5
    Invoke-WebRequest -Uri $hashUrl -OutFile $hashPath -MaximumRedirection 5
} catch {
    Write-Error "❌ Failed to download files. Check your internet connection or URL."
    exit 1
}

Write-Host "🛡️ Verifying integrity..." -ForegroundColor Cyan
if (-not (Test-Path $hashPath)) {
    Write-Error "❌ Hash file not found. Download may have failed."
    exit 1
}

$expectedHash = (Get-Content $hashPath -Raw).Trim()
$actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash

if ($actualHash -ne $expectedHash) {
    Write-Error "❌ Hash mismatch! Expected $expectedHash, got $actualHash. Install aborted."
    Remove-Item $zipPath, $hashPath -ErrorAction SilentlyContinue
    exit 1
}
Write-Host "✅ Integrity verified." -ForegroundColor Green

$installDir = Join-Path ([Environment]::GetFolderPath("UserProfile")) ".jar-cart\bin"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir -Force | Out-Null }

Write-Host "📦 Unpacking to $installDir..." -ForegroundColor Cyan
try {
    if (Test-Path "$installDir\jar-cart.exe") { 
        Remove-Item "$installDir\jar-cart.exe" -Force -ErrorAction SilentlyContinue 
    }
    
    Expand-Archive -Path $zipPath -DestinationPath $installDir -Force
} catch {
    Write-Warning "⚠️ Binary is locked by the OS. Please close all running 'jar-cart' processes and try again."
    exit 1
} finally {
    Remove-Item $zipPath, $hashPath -ErrorAction SilentlyContinue
}

Write-Host "✨ Done! jar-cart $Version is successfully installed." -ForegroundColor Green