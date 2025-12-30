package cmd

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab/config"
	"github.com/nanoteck137/authlab/core"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use: "migrate",
}

var upCmd = &cobra.Command{
	Use: "up",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.LoadedConfig
		conf.RunMigrations = false
		app := core.NewBaseApp(&conf)

		err := app.Bootstrap()
		if err != nil {
			slog.Error("Failed to bootstrap app", "err", err)
			os.Exit(-1)
		}

		err = app.DB().RunMigrateUp()
		if err != nil {
			slog.Error("Failed to run migrate up", "err", err)
			os.Exit(-1)
		}
	},
}

var downCmd = &cobra.Command{
	Use: "down",
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.LoadedConfig
		conf.RunMigrations = false
		app := core.NewBaseApp(&conf)

		err := app.Bootstrap()
		if err != nil {
			slog.Error("Failed to bootstrap app", "err", err)
			os.Exit(-1)
		}

		err = app.DB().RunMigrateDown()
		if err != nil {
			slog.Error("Failed to run migrate down", "err", err)
			os.Exit(-1)
		}
	},
}

// TODO(patrik): Move to dev cmd
var createCmd = &cobra.Command{
	Use:  "create <MIGRATION_NAME>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := goose.Create(nil, "./migrations", name, "sql")
		if err != nil {
			slog.Error("Failed to create migration", "err", err)
			os.Exit(-1)
		}
	},
}

// TODO(patrik): Move to dev cmd?
var fixCmd = &cobra.Command{
	Use: "fix",
	Run: func(cmd *cobra.Command, args []string) {
		err := goose.Fix("./migrations")
		if err != nil {
			slog.Error("Failed to fix migrations", "err", err)
			os.Exit(-1)
		}
	},
}

func init() {
	migrateCmd.AddCommand(upCmd)
	migrateCmd.AddCommand(downCmd)
	migrateCmd.AddCommand(createCmd)
	migrateCmd.AddCommand(fixCmd)

	rootCmd.AddCommand(migrateCmd)
}
