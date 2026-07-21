$packageName = 'jar-cart'
$url = 'https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/v0.5.1/jar-cart-x86_64-windows.zip'
$checksum = '12764C388ACDBA097E1E3EDFB15D181019AED90E250614C05807325B0D2D71ED'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

Install-ChocolateyZipPackage -PackageName $packageName `
                             -Url $url `
                             -UnzipLocation $toolsDir `
                             -Checksum $checksum `
                             -ChecksumType 'sha256'

Install-BinFile -Name 'jar-cart' -Path (Join-Path $toolsDir 'jar-cart.exe')
Install-BinFile -Name 'jc' -Path (Join-Path $toolsDir 'jar-cart.exe')