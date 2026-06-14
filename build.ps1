$ErrorActionPreference = "Stop"

$ldflags = "-s -w -H windowsgui"

Write-Host "Building godot.exe..." -ForegroundColor Cyan
go build -ldflags $ldflags -trimpath -o godot.exe ./cmd/godot
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Get-ChildItem godot.exe | Select-Object Name, @{N='Size (MB)';E={'{0:N2}' -f ($_.Length/1MB)}} | Format-Table -AutoSize
Write-Host "Build complete." -ForegroundColor Green
