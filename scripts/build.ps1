$ErrorActionPreference = "Stop"

if (Test-Path ".\.env") {
    Get-Content ".\.env" | ForEach-Object {
        if ($_.Trim() -and -not $_.StartsWith("#") -and $_.Contains("=")) {
            $key, $value = $_.Trim().Split("=", 2)
            [System.Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim().Trim("'").Trim('"'), "Process")
        }
    }
}

if ($IsWindows) {
    Write-Host "🎨 Generating Windows metadata..." -ForegroundColor Cyan
    if (-not (Get-Command "goversioninfo" -ErrorAction SilentlyContinue)) {
        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
    }
    goversioninfo -manifest=./src/versioninfo.json -out=./src/resource.syso
}

Write-Host "⚡ Compiling optimized jar-cart.exe..." -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -ldflags="-s -w" -trimpath -o jar-cart.exe ./src

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Compilation Successful" -ForegroundColor Green
} else {
    Write-Host "❌ Go Build Failed" -ForegroundColor Red
    exit 1
}

if (Test-Path ".\src\resource.syso") { Remove-Item ".\src\resource.syso" }