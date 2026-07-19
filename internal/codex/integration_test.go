package codex

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallReplacesNotifyAndCreatesBackup(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		t.Fatal(err)
	}
	original := "model = \"gpt-5\"\nnotify = [\"old-notifier\", \"turn-ended\"]\nother = true\n"
	configPath := filepath.Join(codexDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	manager := NewManager(home)
	backup, err := manager.Install(`C:\Apps\codex-bark-notifier.exe`)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if backup.Path == "" || !strings.Contains(backup.Path, "bark-notify-backup") {
		t.Fatalf("unexpected backup: %#v", backup)
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), `notify = ["C:\\Apps\\codex-bark-notifier.exe"]`) {
		t.Fatalf("notify command was not installed: %s", updated)
	}
	backupContent, err := os.ReadFile(backup.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(backupContent) != original {
		t.Fatalf("backup content changed: %q", backupContent)
	}
}

func TestUninstallRestoresOnlyWhenConfigurationWasNotChanged(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(codexDir, "config.toml")
	original := "model = \"gpt-5\"\nnotify = [\"old\"]\n"
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(home)
	if _, err := manager.Install(`C:\Apps\codex-bark-notifier.exe`); err != nil {
		t.Fatal(err)
	}
	if err := manager.Uninstall(); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	restored, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != original {
		t.Fatalf("config was not restored: %q", restored)
	}

	if _, err := manager.Install(`C:\Apps\codex-bark-notifier.exe`); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("notify = [\"someone-else\"]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := manager.Uninstall(); !errors.Is(err, ErrUninstallConflict) {
		t.Fatalf("Uninstall error = %v, want ErrUninstallConflict", err)
	}
}

func TestNotifyLineQuotesWindowsPath(t *testing.T) {
	line := NotifyLine(`C:\Program Files\Codex Bark\codex-bark-notifier.exe`)
	if line != `notify = ["C:\\Program Files\\Codex Bark\\codex-bark-notifier.exe"]` {
		t.Fatalf("NotifyLine = %q", line)
	}
}
