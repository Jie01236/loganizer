package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "loganalyzer",
	Short:         "Analyse des fichiers de logs en parallèle",
	Long:          "loganalyzer permet de charger une configuration JSON et d'analyser plusieurs fichiers de logs en parallèle.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.ExecuteContext(ctx)
}
