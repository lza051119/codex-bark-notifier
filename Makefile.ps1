param(
    [string]$Version = "0.1.0"
)

$ErrorActionPreference = 'Stop'

go fmt ./...
go test ./...
go build -trimpath -ldflags "-H=windowsgui -X github.com/lza051119/codex-bark-notifier/internal/app.Version=$Version" -o "dist/codex-bark-notifier-windows-x64.exe" ./cmd/codex-bark-notifier
