package main

import (
	"log/slog"
	"os"

	"github.com/nanoteck137/authlab/apis"
	"github.com/nanoteck137/pyrin/spark"
	"github.com/nanoteck137/pyrin/spark/typescript"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "internal",
}

var genCmd = &cobra.Command{
	Use: "gen",
	Run: func(cmd *cobra.Command, args []string) {
		router := spark.Router{}
		apis.RegisterHandlers(nil, &router)

		nameFilter := spark.NameFilter{}

		serverDef, err := spark.CreateServerDef(&router, nameFilter)
		if err != nil {
			slog.Error("failed to create server def", "err", err)
			os.Exit(-1)
		}

		err = serverDef.SaveToFile("misc/pyrin.json")
		if err != nil {
			slog.Error("failed save server def", "err", err)
			os.Exit(-1)
		}

		slog.Info("Wrote 'misc/pyrin.json'")

		resolver, err := spark.CreateResolverFromServerDef(&serverDef)
		if err != nil {
			slog.Error("failed to create resolver", "err", err)
			os.Exit(-1)
		}

		{
			gen := typescript.TypescriptGenerator{}

			err = gen.Generate(&serverDef, resolver, "web/src/lib/api")
			if err != nil {
				slog.Error("failed to generate typescript client", "err", err)
				os.Exit(-1)
			}
		}

		// {
		// 	gen := golang.GolangGenerator{}
		//
		// 	err = gen.Generate(&serverDef, resolver, "cmd/authlab-cli/api")
		// 	if err != nil {
		// 		slog.Error("failed to generate golang client", "err", err)
		// 		os.Exit(-1)
		// 	}
		// }
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error("Failed to execute", "err", err)
		os.Exit(-1)
	}
}
