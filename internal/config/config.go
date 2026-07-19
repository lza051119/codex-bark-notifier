package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	configFileName = "config.json"
	defaultServer  = "https://api.day.app"
)

// Settings is the user-facing configuration. DeviceKey is encrypted before
// it is written to disk.
type Settings struct {
	ServerURL string
	Enabled   bool
	DeviceKey string
}

type fileSettings struct {
	ServerURL          string `json:"server_url"`
	Enabled            bool   `json:"enabled"`
	ProtectedDeviceKey string `json:"device_key_protected"`
}

// Store persists settings under one application-data directory.
type Store struct {
	Dir string
}

func NewStore(dir string) *Store {
	if strings.TrimSpace(dir) == "" {
		dir = DefaultDir()
	}
	return &Store{Dir: dir}
}

func DefaultDir() string {
	root, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "CodexBarkNotifier")
	}
	return filepath.Join(root, "CodexBarkNotifier")
}

func (s *Store) path() string {
	return filepath.Join(s.Dir, configFileName)
}

func (s *Store) Load() (Settings, error) {
	raw, err := os.ReadFile(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return Settings{ServerURL: defaultServer}, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read settings: %w", err)
	}

	var stored fileSettings
	if err := json.Unmarshal(raw, &stored); err != nil {
		return Settings{}, errors.New("settings file is invalid")
	}
	deviceKey := ""
	if stored.ProtectedDeviceKey != "" {
		deviceKey, err = UnprotectSecret(stored.ProtectedDeviceKey)
		if err != nil {
			return Settings{}, errors.New("device key could not be decrypted")
		}
	}
	if strings.TrimSpace(stored.ServerURL) == "" {
		stored.ServerURL = defaultServer
	}
	return Settings{ServerURL: stored.ServerURL, Enabled: stored.Enabled, DeviceKey: deviceKey}, nil
}

func (s *Store) Save(settings Settings) error {
	protected := ""
	var err error
	if strings.TrimSpace(settings.DeviceKey) != "" {
		protected, err = ProtectSecret(settings.DeviceKey)
		if err != nil {
			return errors.New("device key could not be protected")
		}
	}
	stored := fileSettings{
		ServerURL:          strings.TrimSpace(settings.ServerURL),
		Enabled:            settings.Enabled,
		ProtectedDeviceKey: protected,
	}
	if stored.ServerURL == "" {
		stored.ServerURL = defaultServer
	}
	raw, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}
	tempName := filepath.Join(s.Dir, ".config.json.tmp")
	if err := os.WriteFile(tempName, append(raw, '\n'), 0600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	if err := os.Rename(tempName, s.path()); err != nil {
		_ = os.Remove(tempName)
		return fmt.Errorf("replace settings: %w", err)
	}
	return nil
}

// ProtectSecret protects data with the current user's Windows DPAPI.
func ProtectSecret(secret string) (string, error) {
	if secret == "" {
		return "", nil
	}
	data := []byte(secret)
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	protected := unsafe.Slice(out.Data, out.Size)
	return base64.StdEncoding.EncodeToString(protected), nil
}

// UnprotectSecret reverses ProtectSecret for the current Windows user.
func UnprotectSecret(protected string) (string, error) {
	if protected == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(protected)
	if err != nil || len(data) == 0 {
		return "", errors.New("invalid protected value")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return string(unsafe.Slice(out.Data, out.Size)), nil
}
