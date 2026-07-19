$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
$userProfile = [Environment]::GetEnvironmentVariable('USERPROFILE')

$tracked = @(git ls-files)
if ($LASTEXITCODE -ne 0) {
    throw 'Could not enumerate tracked files.'
}

$forbiddenNames = @('config.local.json', 'codex-bark-notify.log')
foreach ($path in $tracked) {
    $leaf = Split-Path -Leaf $path
    if ($forbiddenNames -contains $leaf) {
        throw "Forbidden local file is tracked: $path"
    }
    if ($path -match '(^|/)(dist|build|coverage)/') {
        throw "Build output is tracked: $path"
    }
    $full = Join-Path $root $path
    if (-not (Test-Path -LiteralPath $full -PathType Leaf)) {
        continue
    }
    $text = Get-Content -Raw -LiteralPath $full
    if ($userProfile -and $text.Contains($userProfile)) {
        throw "Machine-specific user path found in tracked file: $path"
    }
    if ($text -match '(?i)https?://(?:[^/\s]+\.)?day\.app/[A-Za-z0-9_-]{16,}') {
        throw "Possible live Bark device key found in tracked file: $path"
    }
}

Write-Host "Security scan passed for $($tracked.Count) tracked files."
