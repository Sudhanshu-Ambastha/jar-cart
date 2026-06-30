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

$hash = Get-FileHash -Path $BinaryName -Algorithm SHA256
$hash.Hash | Out-File "checksums.txt" -Encoding utf8

Compress-Archive -Path $BinaryName, "checksums.txt" -DestinationPath $AssetName -Force
Remove-Item "checksums.txt"