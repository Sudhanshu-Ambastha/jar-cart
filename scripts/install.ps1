param(
    [string]$Version = ""
)
$ErrorActionPreference = "Stop"

if ([string]::IsNullOrEmpty($Version)) {
    Write-Host "🔍 Fetching latest version tag from GitHub API..."
    try {
        $releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest" -UseBasicParsing
        $Version = $releaseInfo.tag_name
    } catch {
        Write-Host "⚠️ API check failed. Falling back to default v0.0.1"
        $Version = "v0.0.1"
    }
}

$installDir = Join-Path $HOME ".jar-cart\bin"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Path $installDir | Out-Null }

$url = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$Version/jar-cart-x86_64-windows.zip"
$zipPath = Join-Path [System.IO.Path]::GetTempPath() "jar-cart.zip"

Write-Host "⚡ Downloading jar-cart $Version for Windows..."
Invoke-WebRequest -Uri $url -OutFile $zipPath

Write-Host "📦 Unpacking core workspace tools..."
Expand-Archive -Path $zipPath -DestinationPath $installDir -Force
Remove-Item $zipPath

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "🚀 Automatically added $installDir to your User PATH variable."
}

Write-Host "✨ Successfully installed jar-cart! Restart your shell session and type: jar-cart --help"