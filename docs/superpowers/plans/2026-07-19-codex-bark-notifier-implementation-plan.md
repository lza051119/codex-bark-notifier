# Codex Bark Notifier Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and publish a public Windows x64 Go application that configures Bark notifications, receives Codex task-completion events, and provides a tested GitHub Release.

**Architecture:** A single Go executable has GUI mode and event mode. GUI mode uses native Windows controls through `github.com/lxn/walk`; event mode accepts Codex's legacy notify argument or Stop-hook JSON on stdin, forwards the existing Windows notifier, and sends an HTTPS Bark request. User configuration lives under `%APPDATA%\\CodexBarkNotifier`; the Bark device key is protected with Windows DPAPI.

**Tech Stack:** Go 1.26+, `github.com/lxn/walk`, `github.com/lxn/win`, `golang.org/x/sys/windows`, Go standard library HTTP/JSON/testing, GitHub CLI, Windows PowerShell.

## Global Constraints

- Project root: `D:\\projects\\codex-bark-notifier`.
- Public repository: `lza051119/codex-bark-notifier`.
- Build target: Windows amd64, GUI subsystem (`-H=windowsgui`).
- No current user's Bark key, Codex config, logs, absolute home paths, or event payloads may enter Git.
- Bark delivery failures must never make Codex task completion fail.
- The mobile notification switch controls Bark delivery only; it does not disable Codex or local Windows notifications.
- Every production behavior is introduced test-first: write a failing test, observe the expected failure, implement the minimum behavior, then rerun the test.

---

### Task 1: Bootstrap module, source layout, and safe repository defaults

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `cmd/codex-bark-notifier/main.go`
- Create: `internal/app/version.go`
- Create: `Makefile.ps1`

**Interfaces:**
- The `main` package will call `app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)`.
- `internal/app/version.go` will expose `const Version = "0.1.0"`.

- [ ] **Step 1: Write the failing bootstrap test**

Create `internal/app/app_test.go` with a test that expects `app.ParseMode([]string{})` to return `"gui"` and `app.ParseMode([]string{"--event", "{}"})` to return `"event"`.

- [ ] **Step 2: Run the test and verify it fails**

Run `go test ./internal/app`. Expected: FAIL because `ParseMode` does not yet exist.

- [ ] **Step 3: Add the minimum module and mode implementation**

Set the module path to `github.com/lza051119/codex-bark-notifier`, add the Windows-only dependencies, implement `ParseMode`, and make `main.go` call `app.Run`.

- [ ] **Step 4: Run the test and verify it passes**

Run `go test ./internal/app`; expected: PASS.

- [ ] **Step 5: Add repository hygiene**

Ignore `.env`, `*.key`, `*.log`, local app data, `dist/`, `build/`, and temporary test configuration files. Add a PowerShell build script that runs `gofmt`, `go test ./...`, and a GUI-subsystem build with version metadata.

- [ ] **Step 6: Commit**

```powershell
git add go.mod go.sum .gitignore cmd internal/app Makefile.ps1
git commit -m "build: bootstrap notifier application"
```

### Task 2: Event parsing and notification content

**Files:**
- Create: `internal/event/event.go`
- Create: `internal/event/event_test.go`

**Interfaces:**
- `event.Parse(raw string) (Notification, error)` accepts a Codex legacy `agent-turn-complete` event or a Stop-hook JSON object.
- `event.Notification` contains `Kind`, `ThreadID`, `CWD`, `LastAssistantMessage`, `InputMessages`, and `IsSubagent`.
- `event.BuildContent(Notification) (title, body string)` returns a project-based title and length-limited body.

- [ ] **Step 1: Write failing tests**

Test that a Stop event produces a notification, a `SubagentStop` event is rejected as non-main, legacy `agent-turn-complete` is accepted, child metadata is marked subagent, and content truncates to 180 body characters.

- [ ] **Step 2: Run `go test ./internal/event` and confirm the expected failures**

- [ ] **Step 3: Implement the smallest parser and content builder**

Use `encoding/json`, normalize both hyphenated and underscored field names, detect `parent_thread_id`, `thread_source == "subagent"`, and nested `source.subagent`, and never include the full event in an error string.

- [ ] **Step 4: Rerun the package tests and confirm PASS**

- [ ] **Step 5: Commit**

```powershell
git add internal/event
git commit -m "feat: parse Codex completion events"
```

### Task 3: Protected configuration and Bark HTTP client

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `internal/bark/client.go`
- Create: `internal/bark/client_test.go`

**Interfaces:**
- `config.Settings` contains `ServerURL`, `Enabled`, and an encrypted device-key record.
- `config.Store.Load() (Settings, error)` and `config.Store.Save(Settings) error` read/write `%APPDATA%\\CodexBarkNotifier\\config.json`.
- `config.ProtectSecret` and `config.UnprotectSecret` use Windows DPAPI in production and return errors on failure.
- `bark.Client.Send(ctx, title, body string) error` performs an encoded GET request to the normalized server URL.

- [ ] **Step 1: Write failing tests**

Use a temporary config directory and `httptest.Server` to test round-trip settings, device-key non-plaintext storage, URL normalization, title/body encoding, non-2xx errors, and timeout/cancellation.

- [ ] **Step 2: Run package tests and confirm they fail for missing implementations**

- [ ] **Step 3: Implement DPAPI-backed storage and HTTP delivery**

Use `CryptProtectData`/`CryptUnprotectData` through `golang.org/x/sys/windows`. Trim trailing slashes from the server URL, append the device key and encoded path components, add `level=critical`, `volume=5`, `call=1`, `sound=minuet`, and `group=codex`, and redact secrets from all errors.

- [ ] **Step 4: Rerun tests and confirm PASS**

- [ ] **Step 5: Commit**

```powershell
git add internal/config internal/bark
git commit -m "feat: add protected Bark configuration and client"
```

### Task 4: Codex integration, local notifier forwarding, and event runner

**Files:**
- Create: `internal/codex/integration.go`
- Create: `internal/codex/integration_test.go`
- Create: `internal/runner/runner.go`
- Create: `internal/runner/runner_test.go`

**Interfaces:**
- `codex.Install(exePath string) (Backup, error)` writes the user's `%USERPROFILE%\\.codex\\config.toml` with a `notify` command pointing to the stable installed exe and saves a timestamped backup.
- `codex.Uninstall() error` restores only a backup marked by this app and leaves unrelated current configuration untouched.
- `codex.FindWindowsNotifier() (string, bool)` discovers the newest existing Codex computer-use notifier.
- `runner.Handle(rawArgs []string, stdin io.Reader) error` parses the event, forwards local notification, checks settings, and attempts Bark delivery while returning nil for delivery failures.

- [ ] **Step 1: Write failing tests**

Test installation against a temporary fake Codex home, exact replacement of the `notify` line, backup creation, idempotent install, safe uninstall mismatch, forwarding command construction, disabled delivery, child-event suppression, and Bark failure isolation.

- [ ] **Step 2: Run `go test ./internal/codex ./internal/runner` and confirm expected failures**

- [ ] **Step 3: Implement integration and runner**

The event runner must support an event JSON argument and stdin JSON. It must ignore subagent events, preserve exit code zero for malformed events/missing config/Bark failures, and write only safe timestamped status messages to `%APPDATA%\\CodexBarkNotifier\\codex-bark-notify.log`.

- [ ] **Step 4: Rerun the focused tests and then `go test ./...`**

- [ ] **Step 5: Commit**

```powershell
git add internal/codex internal/runner
git commit -m "feat: connect Codex completion events to Bark"
```

### Task 5: Native Windows configuration window

**Files:**
- Create: `internal/ui/window_windows.go`
- Create: `internal/ui/window_test.go`
- Modify: `cmd/codex-bark-notifier/main.go`

**Interfaces:**
- `ui.Show(store config.Store, codexManager codex.Manager, client bark.Client) error` opens a native window with server URL, device key, enabled checkbox, Save, Test, Install, Uninstall, and README buttons.
- The GUI uses a hidden/non-console executable build and returns user-facing validation errors without exposing the device key.

- [ ] **Step 1: Write failing UI-facing tests**

Test the view-model validation independently of native window creation: invalid URL is rejected, blank key is rejected for enabled/test actions, disabled settings save, and status text never includes the key.

- [ ] **Step 2: Run the UI package tests and confirm expected failures**

- [ ] **Step 3: Implement the minimum native window and view model**

Use Walk native controls, keep business logic in testable view-model functions, and invoke the same config/codex/bark interfaces as event mode. The enabled checkbox changes only the Bark setting.

- [ ] **Step 4: Rerun tests and compile with `GOOS=windows GOARCH=amd64`**

- [ ] **Step 5: Commit**

```powershell
git add internal/ui cmd/codex-bark-notifier/main.go
git commit -m "feat: add native configuration window"
```

### Task 6: Documentation, installer workflow, CI, and release packaging

**Files:**
- Create: `README.md`
- Create: `LICENSE`
- Create: `.github/workflows/release.yml`
- Create: `scripts/build-release.ps1`
- Create: `scripts/security-scan.ps1`
- Modify: `Makefile.ps1`

**Interfaces:**
- `scripts/build-release.ps1` produces `dist/codex-bark-notifier-windows-x64.exe` and `dist/SHA256SUMS.txt`.
- `scripts/security-scan.ps1` fails if tracked files contain known user-specific paths, Bark key-shaped values, or local config/log files.
- README documents Windows installation, App configuration, Codex integration install/uninstall, iPhone Bark installation/configuration, Android/custom-server notes, test notification, privacy, and recovery.

- [ ] **Step 1: Write failing packaging checks**

Add a PowerShell check that expects the release exe and checksum to exist after the build, and a scan test fixture proving a fake key is rejected while placeholders pass.

- [ ] **Step 2: Run the checks before packaging and confirm expected failures**

- [ ] **Step 3: Implement documentation and packaging scripts**

Build with version metadata, run `go test ./...`, run the security scan before archiving, and calculate SHA-256 from the final exe. The GitHub workflow must build on `windows-latest`, run tests, scan secrets, create the exe and checksum, and attach both to a version tag.

- [ ] **Step 4: Rerun packaging checks and inspect README content**

- [ ] **Step 5: Commit**

```powershell
git add README.md LICENSE .github scripts Makefile.ps1
git commit -m "docs: add installation and release packaging"
```

### Task 7: Full verification and public GitHub Release

**Files:**
- Modify only files required by verification fixes.
- Create: `dist/codex-bark-notifier-windows-x64.exe`
- Create: `dist/SHA256SUMS.txt`
- Create: `release-notes.md`

- [ ] **Step 1: Run the full fresh verification**

```powershell
go test ./...
.\scripts\security-scan.ps1
.\scripts\build-release.ps1 -Version 0.1.0
Get-FileHash .\dist\codex-bark-notifier-windows-x64.exe -Algorithm SHA256
```

Expected: all Go tests pass, security scan exits 0, Windows x64 build exits 0, checksum file matches the executable.

- [ ] **Step 2: Inspect scope and commit release artifacts**

```powershell
git status -sb
git diff --check
git add .
git commit -m "release: package Codex Bark notifier v0.1.0"
```

- [ ] **Step 3: Create the public repository and push**

```powershell
gh repo create lza051119/codex-bark-notifier --public --source . --remote origin --push --description "Native Windows Bark notifications for Codex task completion"
```

- [ ] **Step 4: Create the GitHub Release**

```powershell
gh release create v0.1.0 dist/codex-bark-notifier-windows-x64.exe dist/SHA256SUMS.txt --repo lza051119/codex-bark-notifier --title "Codex Bark Notifier v0.1.0" --notes-file release-notes.md
```

The release notes will state that the executable is Windows x64, the app stores the device key with DPAPI, and the public source contains no user-specific key or config.

- [ ] **Step 5: Verify the remote repository and release**

```powershell
gh repo view lza051119/codex-bark-notifier --json nameWithOwner,isPrivate,defaultBranchRef,url
gh release view v0.1.0 --repo lza051119/codex-bark-notifier
```

Expected: `isPrivate` is `false`, the default branch is `main`, and both release assets are listed.
