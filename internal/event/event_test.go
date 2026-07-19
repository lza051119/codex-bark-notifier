package event

import (
	"strings"
	"testing"
)

func TestParseStopEvent(t *testing.T) {
	raw := `{"hook_event_name":"Stop","session_id":"thread-1","cwd":"C:\\work\\demo","last_assistant_message":"已完成"}`
	notification, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if notification.Kind != "Stop" || notification.ThreadID != "thread-1" {
		t.Fatalf("unexpected notification: %#v", notification)
	}
	if notification.IsSubagent {
		t.Fatal("main Stop event was marked as subagent")
	}
}

func TestParseSubagentStopEventKeepsItsKindAndMetadata(t *testing.T) {
	raw := `{"hook_event_name":"SubagentStop","session_id":"child-1","parent_thread_id":"parent-1"}`
	notification, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if notification.Kind != "SubagentStop" || !notification.IsSubagent {
		t.Fatalf("unexpected subagent notification: %#v", notification)
	}
}

func TestParseLegacyCompletionEvent(t *testing.T) {
	raw := `{"type":"agent-turn-complete","thread-id":"thread-2","cwd":"C:\\work\\demo","last-assistant-message":"done","input-messages":["please build"]}`
	notification, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if notification.Kind != "agent-turn-complete" || notification.ThreadID != "thread-2" {
		t.Fatalf("unexpected legacy notification: %#v", notification)
	}
	if len(notification.InputMessages) != 1 || notification.InputMessages[0] != "please build" {
		t.Fatalf("unexpected input messages: %#v", notification.InputMessages)
	}
}

func TestParseDetectsNestedSubagentMetadata(t *testing.T) {
	raw := `{"hook_event_name":"Stop","source":{"subagent":true}}`
	notification, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if !notification.IsSubagent {
		t.Fatal("nested subagent metadata was not detected")
	}
}

func TestBuildContentUsesProjectAndLimitsBody(t *testing.T) {
	notification := Notification{CWD: `C:\work\demo`, LastAssistantMessage: strings.Repeat("a", 220)}
	title, body := BuildContent(notification)
	if title != "Codex - demo" {
		t.Fatalf("title = %q, want Codex - demo", title)
	}
	if len([]rune(body)) != 183 || !strings.HasPrefix(body, "结果：") || !strings.HasSuffix(body, "...") {
		t.Fatalf("body was not limited to 180 content runes: length=%d body=%q", len([]rune(body)), body)
	}
}
