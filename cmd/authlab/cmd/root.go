package cmd

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab"
	"github.com/nanoteck137/authlab/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     authlab.AppName,
	Version: authlab.Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to run root command", "err", err)
		os.Exit(-1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(authlab.VersionTemplate(authlab.AppName))

	cobra.OnInitialize(config.InitConfig)

	rootCmd.PersistentFlags().StringVarP(&config.ConfigFile, "config", "c", "", "Config File")
}
