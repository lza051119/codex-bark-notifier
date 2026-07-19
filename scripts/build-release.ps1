param(
    [string]$Version = "0.1.0"
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if ($Version.StartsWith('v')) {
    $Version = $Version.Substring(1)
}

go fmt ./...
go test ./...
& "$PSScriptRoot\security-scan.ps1"

$dist = Join-Path $root 'dist'
New-Item -ItemType Directory -Force -Path $dist | Out-Null
$exe = Join-Path $dist 'codex-bark-notifier-windows-x64.exe'
if (Test-Path $exe) {
    Remove-Item -LiteralPath $exe -Force
}

$ldflags = "-H=windowsgui -X github.com/lza051119/codex-bark-notifier/internal/app.Version=$Version"
go build -trimpath -ldflags $ldflags -o $exe .\cmd\codex-bark-notifier

$hash = (Get-FileHash -LiteralPath $exe -Algorithm SHA256).Hash.ToLowerInvariant()
Set-Content -LiteralPath (Join-Path $dist 'SHA256SUMS.txt') -Value "$hash  codex-bark-notifier-windows-x64.exe" -Encoding ascii
Write-Host "Built $exe"
Write-Host "SHA256 $hash"
