param(
    [string]$BinaryName,
    [string]$AssetName,
    [string]$CertBase64,
    [string]$CertPassword
)

$ErrorActionPreference = "Stop"

if ($CertBase64 -and $CertPassword) {
    $pfxPath = ".\developer.pfx"
    [System.IO.File]::WriteAllBytes($pfxPath, [System.Convert]::FromBase64String($CertBase64))
    $securePassword = ConvertTo-SecureString $CertPassword -AsPlainText -Force
    $certObject = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $securePassword)
    Set-AuthenticodeSignature -FilePath $BinaryName -Certificate $certObject -HashAlgorithm "SHA256"
    Remove-Item $pfxPath -Force
}

Compress-Archive -Path $BinaryName -DestinationPath $AssetName -Force