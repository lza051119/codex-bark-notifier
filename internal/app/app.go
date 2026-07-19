package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/lza051119/codex-bark-notifier/internal/codex"
	"github.com/lza051119/codex-bark-notifier/internal/config"
	"github.com/lza051119/codex-bark-notifier/internal/runner"
	"github.com/lza051119/codex-bark-notifier/internal/ui"
)

// ParseMode selects the executable's user-facing or Codex event entry point.
func ParseMode(args []string) string {
	for _, arg := range args {
		if arg == "--event" {
			return "event"
		}
		if strings.HasPrefix(strings.TrimSpace(arg), "{") {
			return "event"
		}
	}
	return "gui"
}

// Run is the top-level application entry point for both GUI and Codex event
// modes.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if ParseMode(args) == "event" {
		return runner.Handle(args, stdin, runner.Options{})
	}
	if len(args) > 0 && args[0] == "--version" {
		_, err := fmt.Fprintln(stdout, Version)
		return err
	}
	return ui.Show(config.NewStore(""), codex.NewManager(""))
}
