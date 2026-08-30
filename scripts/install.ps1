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
$OldExePath = [System.IO.Path]::Combine($InstallDir, "gdup.exe.old")
$TempExePath = [System.IO.Path]::Combine($InstallDir, "gdup.exe.tmp")

Try {
    # Remove previous .old file if it exists
    if (Test-Path $OldExePath) {
        Remove-Item -Path $OldExePath -Force -ErrorAction Ignore
    }

    # Download to a temporary file
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempExePath

    # Rename existing executable if it's there (Windows allows renaming running files)
    $renamedOld = $false
    if (Test-Path $ExePath) {
        Rename-Item -Path $ExePath -NewName "gdup.exe.old" -Force
        $renamedOld = $true
    }

    # Move the new downloaded file into place
    Try {
        Rename-Item -Path $TempExePath -NewName "gdup.exe" -Force
    } Catch {
        # Rollback: if we renamed the old one, put it back so the app isn't broken
        if ($renamedOld) {
            Rename-Item -Path $OldExePath -NewName "gdup.exe" -Force -ErrorAction Ignore
        }
        throw $_ # Rethrow to outer catch
    }
} Catch {
    Write-Error "Failed to download or install binary: $_"
    # Cleanup temporary file if it exists
    if (Test-Path $TempExePath) {
        Remove-Item -Path $TempExePath -Force -ErrorAction Ignore
    }
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
