package cmd

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab/apis"
	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/core"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		app := core.NewBaseApp(&config.LoadedConfig)

		err := app.Bootstrap()
		if err != nil {
			slog.Error("Failed to bootstrap app", "err", err)
			os.Exit(-1)
		}

		e, err := apis.Server(app)
		if err != nil {
			slog.Error("Failed to create server", "err", err)
			os.Exit(-1)
		}

		err = e.Start(app.Config().ListenAddr)
		if err != nil {
			slog.Error("Failed to start server", "err", err)
			os.Exit(-1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
