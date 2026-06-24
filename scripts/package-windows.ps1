param(
    [string]$BinaryName,
    [string]$AssetName,
    [string]$CertBase64,
    [string]$CertPassword
)

$ErrorActionPreference = "Stop"

# 1. Handle Conditional Code Signing
if ($CertBase64 -and $CertPassword) {
    Write-Host "🎫 Decoding code signing certificate..."
    $pfxPath = ".\developer.pfx"
    
    # Securely write the binary bytes for the certificate
    [System.IO.File]::WriteAllBytes($pfxPath, [System.Convert]::FromBase64String($CertBase64))
    
    Write-Host "🔏 Signing executable binary asset..."
    $securePassword = ConvertTo-SecureString $CertPassword -AsPlainText -Force
    $certObject = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $securePassword)
    
    $status = Set-AuthenticodeSignature -FilePath $BinaryName -Certificate $certObject -HashAlgorithm "SHA256"
    Remove-Item $pfxPath -Force
    
    Write-Host "✨ Signing status: $($status.Status)"
} else {
    Write-Host "⚠️ Missing release secrets. Binary remains unsigned."
}

# 2. Package into Production ZIP Artifact
Write-Host "📦 Compressing $BinaryName into $AssetName..."
Compress-Archive -Path $BinaryName -DestinationPath $AssetName -Force