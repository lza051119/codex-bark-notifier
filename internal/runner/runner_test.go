package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lza051119/codex-bark-notifier/internal/config"
)

func TestHandleIgnoresSubagentEvent(t *testing.T) {
	called := false
	raw := `{"hook_event_name":"SubagentStop","parent_thread_id":"parent"}`
	err := Handle([]string{"--event", raw}, strings.NewReader(""), Options{
		Send: func(context.Context, string, string) error { called = true; return nil },
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if called {
		t.Fatal("subagent event was sent to Bark")
	}
}

func TestHandleForwardsAndSendsEnabledMainEvent(t *testing.T) {
	var forwarded, title, body string
	store := config.NewStore(t.TempDir())
	if err := store.Save(config.Settings{ServerURL: "https://bark.example.test", Enabled: true, DeviceKey: "key"}); err != nil {
		t.Fatal(err)
	}
	raw := `{"type":"agent-turn-complete","cwd":"C:\\work\\demo","last-assistant-message":"done"}`
	err := Handle([]string{raw}, strings.NewReader(""), Options{
		Store:   store,
		Forward: func(raw string) error { forwarded = raw; return nil },
		Send:    func(_ context.Context, gotTitle, gotBody string) error { title, body = gotTitle, gotBody; return nil },
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if forwarded != raw || title != "Codex - demo" || body != "结果：done" {
		t.Fatalf("forwarded=%q title=%q body=%q", forwarded, title, body)
	}
}

func TestHandleDoesNotFailWhenBarkFails(t *testing.T) {
	store := config.NewStore(t.TempDir())
	if err := store.Save(config.Settings{ServerURL: "https://bark.example.test", Enabled: true, DeviceKey: "key"}); err != nil {
		t.Fatal(err)
	}
	raw := `{"hook_event_name":"Stop","cwd":"C:\\work\\demo","last_assistant_message":"done"}`
	err := Handle([]string{"--event", raw}, strings.NewReader(""), Options{
		Store: store,
		Send:  func(context.Context, string, string) error { return errors.New("network down") },
	})
	if err != nil {
		t.Fatalf("Handle returned Bark failure: %v", err)
	}
}

func TestHandleDoesNotSendWhenDisabled(t *testing.T) {
	called := false
	store := config.NewStore(t.TempDir())
	if err := store.Save(config.Settings{ServerURL: "https://bark.example.test", Enabled: false, DeviceKey: "key"}); err != nil {
		t.Fatal(err)
	}
	raw := `{"hook_event_name":"Stop","cwd":"C:\\work\\demo","last_assistant_message":"done"}`
	if err := Handle([]string{"--event", raw}, strings.NewReader(""), Options{
		Store: store,
		Send:  func(context.Context, string, string) error { called = true; return nil },
	}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("disabled notification was sent")
	}
}
