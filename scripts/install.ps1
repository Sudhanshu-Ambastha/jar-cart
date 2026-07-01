param([string]$Version = "")
$ErrorActionPreference = "Stop"

# 1. Fetch Version
if ([string]::IsNullOrEmpty($Version)) {
    Write-Host "🔍 Fetching latest version from GitHub..." -ForegroundColor Cyan
    try {
        $info = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest"
        $Version = $info.tag_name
    } catch {
        Write-Error "❌ Failed to fetch latest version. Please provide a version manually."
        exit 1
    }
}

$arch = "x86_64"
$baseUrl = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version"
$zipUrl = "$baseUrl/jar-cart-$arch-windows.zip"
$hashUrl = "$baseUrl/checksums.txt"

$tempDir = [System.IO.Path]::GetTempPath()
$zipPath = Join-Path $tempDir "jar-cart.zip"
$hashPath = Join-Path $tempDir "checksums.txt"

# 2. Download Assets
Write-Host "⚡ Downloading $Version..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath
Invoke-WebRequest -Uri $hashUrl -OutFile $hashPath

# 3. Integrity Check
Write-Host "🛡️ Verifying integrity..."
$expectedHash = (Get-Content $hashPath | Select-String "jar-cart-$arch-windows.zip").ToString().Split(' ')[0]
$actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash

if ($actualHash -ne $expectedHash) {
    Write-Error "❌ Hash mismatch! Update aborted. Your existing version remains untouched."
    exit 1
}
Write-Host "✅ Integrity verified." -ForegroundColor Green

# 4. Perform Update (Only now we clean and replace)
$installDir = Join-Path $HOME ".jar-cart\bin"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir -Force }

Write-Host "📦 Unpacking..."
# Clean old files only after successful verification
if (Test-Path "$installDir\jar-cart.exe") { Remove-Item "$installDir\jar-cart.exe" -Force }
if (Test-Path "$installDir\checksums.txt") { Remove-Item "$installDir\checksums.txt" -Force }

Expand-Archive -Path $zipPath -DestinationPath $installDir -Force
Copy-Item $hashPath -Destination (Join-Path $installDir "checksums.txt") -Force

# Cleanup temp files
Remove-Item $zipPath
Remove-Item $hashPath

# 5. Update Path
$oldPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($oldPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$oldPath;$installDir", "User")
}

Write-Host "✨ Done! jar-cart is updated to $Version." -ForegroundColor Green