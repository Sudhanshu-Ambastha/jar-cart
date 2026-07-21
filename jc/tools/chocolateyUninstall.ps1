$packageName = 'jar-cart'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

Uninstall-BinFile -Name 'jar-cart'
Uninstall-BinFile -Name 'jc'

$exePath = Join-Path $toolsDir 'jar-cart.exe'
if (Test-Path $exePath) {
    Remove-Item $exePath -Force
}