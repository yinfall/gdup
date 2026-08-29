$ErrorActionPreference = "Stop"

Set-Location -Path "$PSScriptRoot\.."

$ldflags = "-s -w"
$installDir = [System.IO.Path]::Combine([Environment]::GetFolderPath("UserProfile"), ".gdup", "bin")

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

$output = [System.IO.Path]::Combine($installDir, "gdup.exe")

Write-Host "Building gdup to $installDir..." -ForegroundColor Cyan
go build -ldflags $ldflags -trimpath -o $output ./cmd/gdup
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Get-ChildItem $output | Select-Object Name, @{N='Size (MB)';E={'{0:N2}' -f ($_.Length/1MB)}} | Format-Table -AutoSize
Write-Host "Build and install complete!" -ForegroundColor Green
Write-Host "Make sure to add $installDir to your system PATH." -ForegroundColor Yellow
Write-Host "If you want to use the transparent 'godot' command, run: gdup shim install" -ForegroundColor Cyan
