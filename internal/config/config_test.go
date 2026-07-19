package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreRoundTripsSettingsWithoutWritingPlaintextKey(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	want := Settings{ServerURL: "https://bark.example.test/", Enabled: true, DeviceKey: "secret-device-key"}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, configFileName))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if strings.Contains(string(raw), want.DeviceKey) {
		t.Fatal("config file contains the plaintext device key")
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got != want {
		t.Fatalf("Load = %#v, want %#v", got, want)
	}
}

func TestProtectAndUnprotectSecretRoundTrip(t *testing.T) {
	protected, err := ProtectSecret("secret-device-key")
	if err != nil {
		t.Fatalf("ProtectSecret returned error: %v", err)
	}
	if protected == "secret-device-key" || protected == "" {
		t.Fatalf("ProtectSecret returned unsafe value %q", protected)
	}
	got, err := UnprotectSecret(protected)
	if err != nil {
		t.Fatalf("UnprotectSecret returned error: %v", err)
	}
	if got != "secret-device-key" {
		t.Fatalf("UnprotectSecret = %q, want original secret", got)
	}
}
