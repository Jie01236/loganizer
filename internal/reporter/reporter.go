package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"loganalyzer/internal/analyzer"
)

// Export ecrit the analyses a un JSON file.
func Export(path string, results []analyzer.LogAnalysisResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("création du dossier de sortie %q: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("sérialisation du rapport JSON: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("écriture du fichier de rapport %q: %w", path, err)
	}

	return nil
}
