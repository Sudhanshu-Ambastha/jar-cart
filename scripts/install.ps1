param([string]$Version = "")
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrEmpty($Version)) {
    Write-Host "🔍 Fetching latest version from GitHub..." -ForegroundColor Cyan
    try {
        $info = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest"
        $Version = $info.tag_name
    } catch {
        Write-Error "❌ Failed to fetch latest version."
        exit 1
    }
}

$arch = "x86_64"
$baseUrl = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version"
$zipName = "jar-cart-$arch-windows.zip"
$zipUrl = "$baseUrl/$zipName"
$hashUrl = "$zipUrl.sha256" 

$tempDir = [System.IO.Path]::GetTempPath()
$zipPath = Join-Path $tempDir $zipName
$hashPath = Join-Path $tempDir "$zipName.sha256"

Write-Host "⚡ Downloading $Version..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath
Invoke-WebRequest -Uri $hashUrl -OutFile $hashPath

Write-Host "🛡️ Verifying integrity..."
$expectedHash = (Get-Content $hashPath).ToString().Trim()
$actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash

if ($actualHash -ne $expectedHash) {
    Write-Error "❌ Hash mismatch! Update aborted."
    exit 1
}
Write-Host "✅ Integrity verified." -ForegroundColor Green
$installDir = Join-Path $HOME ".jar-cart\bin"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir -Force }

Write-Host "📦 Unpacking..."
if (Test-Path "$installDir\jar-cart.exe") { Remove-Item "$installDir\jar-cart.exe" -Force }

Expand-Archive -Path $zipPath -DestinationPath $installDir -Force
Remove-Item $zipPath
Remove-Item $hashPath

Write-Host "✨ Done! jar-cart is updated to $Version." -ForegroundColor Green