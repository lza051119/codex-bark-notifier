package codex

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrUninstallConflict = errors.New("Codex configuration changed after installation")

type Backup struct {
	Path string
}

type installManifest struct {
	BackupPath   string `json:"backup_path"`
	ExpectedLine string `json:"expected_line"`
}

type Manager struct {
	Home string
}

func NewManager(home string) *Manager {
	if strings.TrimSpace(home) == "" {
		home, _ = os.UserHomeDir()
	}
	return &Manager{Home: home}
}

func (m *Manager) codexDir() string {
	return filepath.Join(m.Home, ".codex")
}

func (m *Manager) configPath() string {
	return filepath.Join(m.codexDir(), "config.toml")
}

func (m *Manager) manifestPath() string {
	return filepath.Join(m.codexDir(), "codex-bark-notify-install.json")
}

func NotifyLine(exePath string) string {
	return "notify = [" + strconv.Quote(filepath.Clean(exePath)) + "]"
}

func (m *Manager) Install(exePath string) (Backup, error) {
	configPath := m.configPath()
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return Backup{}, fmt.Errorf("read Codex config: %w", err)
	}
	manifest := installManifest{}
	if existing, readErr := os.ReadFile(m.manifestPath()); readErr == nil {
		_ = json.Unmarshal(existing, &manifest)
	}
	if manifest.BackupPath == "" {
		manifest.BackupPath = filepath.Join(m.codexDir(), "config.toml.bark-notify-backup-"+time.Now().Format("20060102-150405")+".toml")
		if err := os.WriteFile(manifest.BackupPath, raw, 0600); err != nil {
			return Backup{}, fmt.Errorf("write Codex backup: %w", err)
		}
	}
	line := NotifyLine(exePath)
	updated := replaceNotifyLine(string(raw), line)
	if err := os.WriteFile(configPath, []byte(updated), 0600); err != nil {
		return Backup{}, fmt.Errorf("write Codex config: %w", err)
	}
	manifest.ExpectedLine = line
	manifestRaw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Backup{}, fmt.Errorf("encode integration manifest: %w", err)
	}
	if err := os.WriteFile(m.manifestPath(), append(manifestRaw, '\n'), 0600); err != nil {
		return Backup{}, fmt.Errorf("write integration manifest: %w", err)
	}
	return Backup{Path: manifest.BackupPath}, nil
}

func (m *Manager) Uninstall() error {
	manifestRaw, err := os.ReadFile(m.manifestPath())
	if err != nil {
		return fmt.Errorf("read integration manifest: %w", err)
	}
	var manifest installManifest
	if err := json.Unmarshal(manifestRaw, &manifest); err != nil || manifest.BackupPath == "" || manifest.ExpectedLine == "" {
		return errors.New("integration manifest is invalid")
	}
	current, err := os.ReadFile(m.configPath())
	if err != nil {
		return fmt.Errorf("read Codex config: %w", err)
	}
	if !hasExactLine(string(current), manifest.ExpectedLine) {
		return ErrUninstallConflict
	}
	backup, err := os.ReadFile(manifest.BackupPath)
	if err != nil {
		return fmt.Errorf("read Codex backup: %w", err)
	}
	if err := os.WriteFile(m.configPath(), backup, 0600); err != nil {
		return fmt.Errorf("restore Codex config: %w", err)
	}
	_ = os.Remove(m.manifestPath())
	return nil
}

func replaceNotifyLine(raw, replacement string) string {
	lines := strings.SplitAfter(raw, "\n")
	found := false
	for i, line := range lines {
		withoutNewline := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if strings.HasPrefix(strings.TrimSpace(withoutNewline), "notify") && strings.Contains(withoutNewline, "=") {
			ending := ""
			if strings.HasSuffix(line, "\r\n") {
				ending = "\r\n"
			} else if strings.HasSuffix(line, "\n") {
				ending = "\n"
			}
			lines[i] = replacement + ending
			found = true
		}
	}
	if !found {
		if raw != "" && !strings.HasSuffix(raw, "\n") {
			raw += "\n"
		}
		return raw + replacement + "\n"
	}
	return strings.Join(lines, "")
}

func hasExactLine(raw, expected string) bool {
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(strings.TrimSuffix(line, "\r")) == expected {
			return true
		}
	}
	return false
}

func FindWindowsNotifier() (string, bool) {
	return FindWindowsNotifierIn(filepath.Join(os.Getenv("LOCALAPPDATA"), "OpenAI", "Codex", "runtimes", "cua_node"))
}

func FindWindowsNotifierIn(runtimeRoot string) (string, bool) {
	entries, err := os.ReadDir(runtimeRoot)
	if err != nil {
		return "", false
	}
	type candidate struct {
		path string
		when int64
	}
	var candidates []candidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runtimePath := filepath.Join(runtimeRoot, entry.Name())
		info, err := os.Stat(runtimePath)
		if err != nil {
			continue
		}
		candidatePath := filepath.Join(runtimePath, "bin", "node_modules", "@oai", "sky", "bin", "windows", "codex-computer-use.exe")
		if _, err := os.Stat(candidatePath); err == nil {
			candidates = append(candidates, candidate{path: candidatePath, when: info.ModTime().UnixNano()})
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].when > candidates[j].when })
	if len(candidates) == 0 {
		return "", false
	}
	return candidates[0].path, true
}

func ForwardLocalNotifier(raw string) error {
	notifier, ok := FindWindowsNotifier()
	if !ok {
		return nil
	}
	return exec.Command(notifier, "turn-ended", raw).Run()
}

func configHash(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
