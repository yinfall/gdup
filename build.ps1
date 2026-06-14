$ErrorActionPreference = "Stop"

$ldflags = "-s -w"
$buildFlags = @("-ldflags", $ldflags, "-trimpath")

Write-Host "Building gvm.exe..." -ForegroundColor Cyan
go build @buildFlags -o gvm.exe ./cmd/gvm
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Building godot.exe..." -ForegroundColor Cyan
$godotLdflags = "-s -w -H windowsgui"
go build -ldflags $godotLdflags -trimpath -o godot.exe ./cmd/godot
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Get-ChildItem gvm.exe, godot.exe | Select-Object Name, @{N='Size (MB)';E={'{0:N2}' -f ($_.Length/1MB)}} | Format-Table -AutoSize
Write-Host "Build complete." -ForegroundColor Green
