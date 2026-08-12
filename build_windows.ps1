$ErrorActionPreference = "Stop"

$ldflags = "-s -w"
$output = "godot.exe"

Write-Host "Building $output..." -ForegroundColor Cyan
go build -ldflags $ldflags -trimpath -o $output ./cmd/godot
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Get-ChildItem $output | Select-Object Name, @{N='Size (MB)';E={'{0:N2}' -f ($_.Length/1MB)}} | Format-Table -AutoSize
Write-Host "Build complete." -ForegroundColor Green
