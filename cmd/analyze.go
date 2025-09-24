package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"loganalyzer/internal/analyzer"
	"loganalyzer/internal/config"
	"loganalyzer/internal/reporter"
)

var (
	configPath  string
	outputPath  string
	statusValue string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyse les logs listés dans un fichier de configuration JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(configPath) == "" {
			return fmt.Errorf("le drapeau --config est requis")
		}

		entries, err := config.Load(configPath)
		if err != nil {
			if errors.Is(err, config.ErrConfigParse) {
				return err
			}
			return fmt.Errorf("échec du chargement de la configuration: %w", err)
		}

		results := analyzer.Analyze(cmd.Context(), entries)

		normalizedFilter := strings.ToUpper(strings.TrimSpace(statusValue))
		if normalizedFilter != "" && normalizedFilter != analyzer.StatusOK && normalizedFilter != analyzer.StatusFailed {
			return fmt.Errorf("statut invalide %q: utiliser OK ou FAILED", normalizedFilter)
		}

		var filtered []analyzer.LogAnalysisResult
		if normalizedFilter != "" {
			filtered = analyzer.FilterByStatus(results, normalizedFilter)
		} else {
			filtered = results
		}

		writer := cmd.OutOrStdout()
		for _, res := range filtered {
			if res.ErrorDetails != "" {
				fmt.Fprintf(writer, "[%s] %s (%s) -> %s | %s\n", res.Status, res.LogID, res.FilePath, res.Message, res.ErrorDetails)
			} else {
				fmt.Fprintf(writer, "[%s] %s (%s) -> %s\n", res.Status, res.LogID, res.FilePath, res.Message)
			}
		}

		if normalizedFilter != "" && len(filtered) == 0 {
			fmt.Fprintf(writer, "Aucun résultat avec le statut %s\n", normalizedFilter)
		}

		if strings.TrimSpace(outputPath) != "" {
			target := buildTimestampedPath(outputPath, time.Now())
			if err := reporter.Export(target, filtered); err != nil {
				return err
			}
			fmt.Fprintf(writer, "Rapport exporté vers %s\n", target)
		}

		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&configPath, "config", "c", "", "Chemin du fichier de configuration JSON")
	analyzeCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Chemin du rapport JSON à générer")
	analyzeCmd.Flags().StringVar(&statusValue, "status", "", "Filtrer les résultats par statut (OK ou FAILED)")

	rootCmd.AddCommand(analyzeCmd)
}

func buildTimestampedPath(rawPath string, now time.Time) string {
	cleaned := strings.TrimSpace(rawPath)
	if cleaned == "" {
		return cleaned
	}

	dir := filepath.Dir(cleaned)
	name := filepath.Base(cleaned)
	stamp := now.Format("060102")

	fileName := fmt.Sprintf("%s_%s", stamp, name)
	if dir == "." {
		return fileName
	}

	return filepath.Join(dir, fileName)
}
