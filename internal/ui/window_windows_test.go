//go:build windows

package ui

import (
	"runtime"
	"testing"

	"github.com/lxn/win"
	"github.com/lza051119/codex-bark-notifier/internal/config"
)

func TestNativeWindowCreatesRequiredControls(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ui := &nativeWindow{}
	if err := ui.create(config.Settings{ServerURL: "https://api.day.app", DeviceKey: "test-key", Enabled: true}); err != nil {
		t.Fatalf("native window creation failed: %v", err)
	}
	defer win.DestroyWindow(ui.hwnd)
	win.ShowWindow(ui.hwnd, win.SW_HIDE)

	for name, hwnd := range map[string]win.HWND{
		"URL input":        ui.editURL,
		"device key input": ui.editKey,
		"enabled checkbox": ui.check,
		"status label":     ui.status,
	} {
		if hwnd == 0 {
			t.Errorf("%s was not created", name)
		}
	}
}
