package authlab

import (
	"fmt"
	"log/slog"

	"github.com/nanoteck137/pyrin/trail"
)

var AppName = "authlab"
var CliAppName = AppName + "-cli"

var Version = "no-version"
var Commit = "no-commit"

func VersionTemplate(appName string) string {
	return fmt.Sprintf(
		"%s: %s (%s)\n",
		appName, Version, Commit)
}

func DefaultLogger() *trail.Logger {
	return trail.NewLogger(&trail.Options{
		Debug: Commit == "no-commit",
	})
}

func init() {
	logger := DefaultLogger()
	slog.SetDefault(logger.Logger)
}
