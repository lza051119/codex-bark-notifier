package app

import "testing"

func TestParseModeDefaultsToGUI(t *testing.T) {
	if got := ParseMode(nil); got != "gui" {
		t.Fatalf("ParseMode(nil) = %q, want gui", got)
	}
}

func TestParseModeRecognizesEvent(t *testing.T) {
	if got := ParseMode([]string{"--event", "{}"}); got != "event" {
		t.Fatalf("ParseMode(event args) = %q, want event", got)
	}
}

func TestParseModeRecognizesLegacyJSONArgument(t *testing.T) {
	if got := ParseMode([]string{`{"type":"agent-turn-complete"}`}); got != "event" {
		t.Fatalf("ParseMode(legacy JSON) = %q, want event", got)
	}
}
