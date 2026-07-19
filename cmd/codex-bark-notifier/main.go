package main

import (
	"os"

	"github.com/lza051119/codex-bark-notifier/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
