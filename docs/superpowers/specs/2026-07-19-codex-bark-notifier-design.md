# Codex Bark Notifier: Windows Application Design

## Goal

Deliver a public, Windows-native Go application that receives Codex task-completion notifications and sends a configurable Bark push notification without requiring users to install Node.js. The application must provide an in-app configuration and enable/disable switch, preserve Codex's normal notification behavior, and publish a reproducible Windows x64 release.

## Scope

The first release includes:

- A native Windows x64 executable built with Go.
- A small configuration window for Bark server URL, device key, enabled state, test notification, and Codex integration install/uninstall.
- A notification entry point that accepts Codex's task-completion event and supports the existing Stop-hook JSON shape for compatibility.
- Automatic discovery and forwarding of Codex's existing Windows notifier where possible.
- A user-local configuration file with the Bark device key protected using Windows DPAPI rather than committed or logged in plaintext.
- Backup and safe update of the user's Codex configuration during integration installation.
- Unit tests and integration-style tests for event parsing, configuration, enable/disable behavior, secret handling, and failure isolation.
- Public GitHub repository `lza051119/codex-bark-notifier` and a GitHub Release containing the executable, documentation, and SHA-256 checksums.

Out of scope for v1:

- A macOS or Linux desktop UI.
- Cloud-hosted configuration or telemetry.
- Automatic update infrastructure beyond the GitHub Release download.
- Uploading the current machine's Codex configuration, logs, or Bark key.

## User experience

The app starts with a configuration window. The user can enter:

- Bark server URL, defaulting to `https://api.day.app`.
- Bark device key.
- Mobile notification enabled/disabled state.

The window provides buttons for saving configuration, sending a test notification, installing Codex integration, uninstalling it, and opening the README. The enabled switch controls mobile Bark delivery only; it does not disable Codex itself. Local Windows notification forwarding remains independent.

The install action copies or registers the executable at a stable per-user path, backs up the existing Codex configuration, and configures Codex to invoke the executable on task completion. Uninstall restores the backed-up configuration when the backup matches the integration that the app installed.

## Architecture and data flow

The single executable has two modes:

1. GUI mode: launched by the user for setup and status management.
2. Event mode: launched by Codex with a legacy notify argument or with a JSON event on standard input.

Event flow:

```text
Codex task completes
  -> Codex invokes codex-bark-notifier.exe
  -> executable parses task-completion/Stop event
  -> executable ignores child-agent events
  -> executable forwards the existing local Windows notifier
  -> executable checks enabled state and reads protected Bark settings
  -> executable sends HTTPS Bark request
  -> executable exits successfully even when Bark delivery fails
```

The Bark base URL is normalized and combined with the device key, title, and body. The app accepts a custom Bark-compatible server URL while preventing path/query injection through URL construction and encoding.

## Configuration and security

Configuration is stored under the current user's application data directory, not in the repository. Non-secret settings may be stored as JSON. The device key is protected with Windows DPAPI tied to the current user. Logs contain timestamps and non-sensitive outcome/status messages only; they never contain the key or complete event payload.

The repository will include only placeholders and examples. `.gitignore` will exclude local configuration, logs, build output, and test secrets. CI and release checks will scan tracked files for obvious key patterns before publishing.

## Error handling

- Missing configuration: show a clear GUI validation message; event mode logs a safe reason and exits zero.
- Disabled mobile notifications: skip Bark delivery and exit zero.
- Bark HTTP or DNS failure: log a safe failure status and never block Codex completion.
- Invalid event JSON: log and exit zero.
- Missing or malformed Codex config during install: refuse to overwrite it and provide a backup/manual action.
- Uninstall without a matching backup: leave the user's current configuration untouched.

## Testing and release verification

Tests will cover event shapes, child-session filtering, URL construction, DPAPI/config behavior, disabled/missing-key paths, HTTP success/failure, and idempotent install/uninstall behavior. Release verification will build a Windows x64 executable, run the test suite, inspect the executable metadata, calculate SHA-256, and verify that the release artifact and source contain no user-specific secret or machine path.

## Release contents

The first public release will include:

- `codex-bark-notifier-windows-x64.exe`
- `SHA256SUMS.txt`
- `README.md` with Windows installation steps and phone-side Bark installation/configuration instructions
- Source code in the public repository and GitHub-generated source archives

