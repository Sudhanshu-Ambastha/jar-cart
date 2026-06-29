$ErrorActionPreference = "Stop"

if (Test-Path ".\.env") {
    Get-Content ".\.env" | ForEach-Object {
        if ($_.Trim() -and -not $_.StartsWith("#") -and $_.Contains("=")) {
            $key, $value = $_.Trim().Split("=", 2)
            [System.Environment]::SetEnvironmentVariable($key.Trim(), $value.Trim().Trim("'").Trim('"'), "Process")
        }
    }
}

Write-Host "⚡ Compiling optimized jar-cart.exe..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -trimpath -o jar-cart.exe ./src

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Compilation Successful" -ForegroundColor Green
} else {
    Write-Host "❌ Go Build Failed" -ForegroundColor Red
    exit 1
}

$pfxPath = ".\developer.pfx"
$password = if ($env:JC_CERT_PASSWORD) { $env:JC_CERT_PASSWORD } else { "JarCart123!" }

if (-not (Test-Path $pfxPath)) {
    Write-Host "🎫 Generating local signing certificate..." -ForegroundColor Yellow
    $cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject "CN=Sudhanshu Developer" -CertStoreLocation "Cert:\CurrentUser\My"
    Export-PfxCertificate -Cert $cert -FilePath $pfxPath -Password (ConvertTo-SecureString $password -AsPlainText -Force)
    Remove-Item $cert.PSPath
}

Write-Host "🔏 Signing jar-cart.exe..." -ForegroundColor Cyan
$certObject = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $password)
$status = Set-AuthenticodeSignature -FilePath "jar-cart.exe" -Certificate $certObject -HashAlgorithm "SHA256"

if ($status.Status -eq "Valid") {
    Write-Host "✨ Successfully built and signed jar-cart.exe!" -ForegroundColor Green
} else {
    Write-Host "⚠️ Signing status: $($status.StatusMessage)" -ForegroundColor Yellow
}