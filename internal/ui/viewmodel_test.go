package ui

import (
	"strings"
	"testing"
)

func TestValidateSettingsRejectsInvalidServerURL(t *testing.T) {
	if err := ValidateSettings("not-a-url", "key", true); err == nil {
		t.Fatal("ValidateSettings accepted an invalid server URL")
	}
}

func TestValidateSettingsRequiresKeyWhenEnabled(t *testing.T) {
	if err := ValidateSettings("https://api.day.app", "", true); err == nil {
		t.Fatal("ValidateSettings accepted an enabled configuration without a key")
	}
}

func TestValidateSettingsAllowsBlankKeyWhenDisabled(t *testing.T) {
	if err := ValidateSettings("https://api.day.app", "", false); err != nil {
		t.Fatalf("ValidateSettings rejected disabled configuration: %v", err)
	}
}

func TestStatusTextNeverIncludesDeviceKey(t *testing.T) {
	key := "private-device-key"
	status := StatusText(true, key)
	if strings.Contains(status, key) {
		t.Fatal("status text contains the device key")
	}
}
