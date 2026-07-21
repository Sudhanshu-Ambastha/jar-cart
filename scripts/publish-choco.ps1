param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    [Parameter(Mandatory=$true)]
    [string]$Sha256,
    [Parameter(Mandatory=$true)]
    [string]$ApiKey,
    [Parameter(Mandatory=$true)]
    [string]$TagRef
)

$url = "https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/$TagRef/jar-cart-x86_64-windows.zip"
$nuspecPath = "./jc/jc.nuspec"
$installPath = "./jc/tools/chocolateyinstall.ps1"

if (Test-Path $nuspecPath) {
    $content = Get-Content $nuspecPath -Raw
    $content = $content -replace '<version>.*?</version>', "<version>$Version</version>"
    Set-Content -Path $nuspecPath -Value $content -NoNewline
}

if (Test-Path $installPath) {
    $content = Get-Content $installPath -Raw
    $content = $content -replace "url\s*=\s*['`"][^'`"]*['`"]", "url = '$url'"
    $content = $content -replace "checksum\s*=\s*['`"][^'`"]*['`"]", "checksum = '$Sha256'"
    Set-Content -Path $installPath -Value $content -NoNewline
}

Set-Location ./jc
choco pack
$nupkg = Get-ChildItem "*.nupkg" | Select-Object -ExpandProperty Name
choco push $nupkg --source="'https://push.chocolatey.org/'" --api-key="$ApiKey"