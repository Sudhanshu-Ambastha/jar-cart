$envPath = ".\.env"
if (Test-Path $envPath) {
    Get-Content $envPath | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
            $key, $value = $line.Split("=", 2)
            $value = $value.Trim().Trim("'").Trim('"')
            [System.Environment]::SetEnvironmentVariable($key.Trim(), $value, "Process")
        }
    }
}

echo "⚡ Compiling jar-cart.exe..."
go build -o jar-cart.exe ./src
if ($LASTEXITCODE -ne 0) { 
    echo "❌ Go Build Failed"
    exit 
}

$pfxPath = Resolve-Path ".\developer.pfx" -ErrorAction SilentlyContinue
if (-not $pfxPath) { $pfxPath = ".\developer.pfx" }

$password = $env:JC_CERT_PASSWORD
if (-not $password) {
    $password = "JarCart123!"
}

if (-not (Test-Path $pfxPath)) {
    echo "🎫 Creating a local developer certificate file (developer.pfx)..."
    $securePassword = ConvertTo-SecureString $password -AsPlainText -Force
    $cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject "CN=Sudhanshu Developer" -CertStoreLocation "Cert:\CurrentUser\My" -FriendlyName "JarCartLocalCert" -NotAfter (Get-Date).AddYears(5)
    Export-PfxCertificate -Cert $cert -FilePath $pfxPath -Password $securePassword
    Remove-Item $cert.PSPath
}

$certObject = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $password)

echo "🔏 Signing jar-cart.exe using local pfx..."
$status = Set-AuthenticodeSignature -FilePath "jar-cart.exe" -Certificate $certObject -HashAlgorithm "SHA256"

if ($status.Status -eq "Valid" -or $status.Status -eq "UnknownError") {
    echo "✨ Successfully compiled and signed jar-cart.exe locally!"
} else {
    echo "⚠️ Signing status: $($status.StatusMessage)"
}