$ErrorActionPreference = "Stop"

Set-Location -Path "$PSScriptRoot\.."

$ldflags = "-s -w"
$installDir = [System.IO.Path]::Combine([Environment]::GetFolderPath("UserProfile"), ".gdup", "bin")

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

$output = [System.IO.Path]::Combine($installDir, "gdup.exe")
$outputOld = [System.IO.Path]::Combine($installDir, "gdup.exe.old")
$outputTmp = [System.IO.Path]::Combine($installDir, "gdup.exe.tmp")

Write-Host "Building gdup to $installDir..." -ForegroundColor Cyan
go build -ldflags $ldflags -trimpath -o $outputTmp ./cmd/gdup
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# Prevent 'File in use' error if gdup is currently running
if (Test-Path $outputOld) { Remove-Item -Path $outputOld -Force -ErrorAction Ignore }
if (Test-Path $output) { Rename-Item -Path $output -NewName "gdup.exe.old" -Force -ErrorAction Ignore }
Rename-Item -Path $outputTmp -NewName "gdup.exe" -Force -ErrorAction Stop

Write-Host ""
Get-ChildItem $output | Select-Object Name, @{N='Size (MB)';E={'{0:N2}' -f ($_.Length/1MB)}} | Format-Table -AutoSize
Write-Host "Build and install complete!" -ForegroundColor Green
Write-Host "Make sure to add $installDir to your system PATH." -ForegroundColor Yellow
Write-Host "If you want to use the transparent 'godot' command, run: gdup shim install" -ForegroundColor Cyan
