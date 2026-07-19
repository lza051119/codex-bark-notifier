package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lza051119/codex-bark-notifier/internal/bark"
	"github.com/lza051119/codex-bark-notifier/internal/codex"
	"github.com/lza051119/codex-bark-notifier/internal/config"
	"github.com/lza051119/codex-bark-notifier/internal/event"
)

type Options struct {
	Store   *config.Store
	Forward func(string) error
	Send    func(context.Context, string, string) error
	Log     io.Writer
}

func Handle(rawArgs []string, stdin io.Reader, options Options) error {
	raw := eventPayload(rawArgs, stdin)
	if strings.TrimSpace(raw) == "" {
		logStatus(options, "event input was empty")
		return nil
	}
	notification, err := event.Parse(raw)
	if err != nil {
		logStatus(options, err.Error())
		return nil
	}
	if notification.IsSubagent {
		logStatus(options, "subagent notification ignored")
		return nil
	}

	forward := options.Forward
	if forward == nil {
		forward = codex.ForwardLocalNotifier
	}
	if err := forward(raw); err != nil {
		logStatus(options, "local notification forwarding failed")
	}

	store := options.Store
	if store == nil {
		store = config.NewStore("")
	}
	settings, err := store.Load()
	if err != nil {
		logStatus(options, "settings could not be loaded")
		return nil
	}
	if !settings.Enabled {
		logStatus(options, "Bark notification skipped because mobile alerts are disabled")
		return nil
	}
	if strings.TrimSpace(settings.DeviceKey) == "" {
		logStatus(options, "Bark device key is missing")
		return nil
	}

	title, body := event.BuildContent(notification)
	send := options.Send
	if send == nil {
		client := bark.NewClient(settings.ServerURL, settings.DeviceKey)
		send = client.Send
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := send(ctx, title, body); err != nil {
		logStatus(options, "Bark notification failed")
	}
	return nil
}

func eventPayload(args []string, stdin io.Reader) string {
	if len(args) > 0 {
		if args[0] == "--event" && len(args) > 1 {
			return args[1]
		}
		return args[0]
	}
	raw, _ := io.ReadAll(stdin)
	return string(raw)
}

func logStatus(options Options, message string) {
	writer := options.Log
	if writer == nil {
		dir := config.DefaultDir()
		if os.MkdirAll(dir, 0700) == nil {
			if file, err := os.OpenFile(filepath.Join(dir, "codex-bark-notify.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600); err == nil {
				defer file.Close()
				writer = file
			}
		}
	}
	if writer != nil {
		_, _ = fmt.Fprintf(writer, "%s %s\r\n", time.Now().UTC().Format(time.RFC3339), message)
	}
}
