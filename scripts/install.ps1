$ErrorActionPreference = "Stop"

$Repo = "yinfall/gdup"
$InstallDir = [System.IO.Path]::Combine([Environment]::GetFolderPath("UserProfile"), ".gdup", "bin")

Write-Host "Fetching latest release version from github.com/$Repo..." -ForegroundColor Cyan
$ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
Try {
    $ReleaseInfo = Invoke-RestMethod -Uri $ReleaseUrl
    $LatestTag = $ReleaseInfo.tag_name
} Catch {
    Write-Error "Failed to fetch latest release. Check if the repo is public and has a release."
    exit 1
}

$Binary = "gdup-windows-amd64.exe"
$DownloadUrl = "https://github.com/$Repo/releases/download/$LatestTag/$Binary"

Write-Host "Downloading gdup $LatestTag for Windows..." -ForegroundColor Cyan
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$ExePath = [System.IO.Path]::Combine($InstallDir, "gdup.exe")

Try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExePath
} Catch {
    Write-Error "Failed to download binary from $DownloadUrl"
    exit 1
}

Write-Host "
==========================================================" -ForegroundColor Green
Write-Host " Success! gdup has been installed to:" -ForegroundColor Green
Write-Host "   $InstallDir" -ForegroundColor White
Write-Host ""
Write-Host " PLEASE ADD THIS DIRECTORY TO YOUR SYSTEM PATH!" -ForegroundColor Yellow
Write-Host "==========================================================
" -ForegroundColor Green
Write-Host "After updating your PATH, open a new terminal and run 'gdup help'."
