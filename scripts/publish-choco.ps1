param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    [Parameter(Mandatory=$true)]
    [string]$ApiKey
)

$packageName = "jar-cart"
Write-Host "Checking if version $Version already exists or is in review on Chocolatey..."
$checkUrl = "https://community.chocolatey.org/api/v2/Packages(`$id='$packageName',`$version='$Version')"
$skipPush = $false

try {
    $response = Invoke-WebRequest -Uri $checkUrl -Method Get -UseBasicParsing -ErrorAction Stop
    if ($response.StatusCode -eq 200) {
        Write-Host "::notice::Version $Version already exists on Chocolatey. Skipping push."
        $skipPush = $true
    }
} catch {
    Write-Host "Version $Version not found on live feed. Proceeding..."
}

if (-Not $skipPush) {
    Write-Host "Cleaning up old .nupkg files..."
    Get-ChildItem -Path "./jc" -Filter "*.nupkg" | Remove-Item -Force
    Get-ChildItem -Filter "*.nupkg" | Remove-Item -Force

    Write-Host "Building new .nupkg package..."
    Set-Location ./jc
    choco pack

    $nupkgName = "jar-cart.$Version.nupkg"
    if (-Not (Test-Path $nupkgName)) {
        throw "Expected package $nupkgName was not found after running choco pack."
    }

    try {
        Write-Host "Registering API key and pushing to Chocolatey..."
        choco apikey --key "$ApiKey" --source "https://push.chocolatey.org/"
        choco push "./$nupkgName" --source "https://push.chocolatey.org/"
    } catch {
        Write-Warning "Chocolatey push failed (possibly due to 403 or pending moderation queue). Metadata changes are already safely committed to GitHub."
        Write-Warning $_.Exception.Message
    } finally {
        Set-Location ..
    }
}