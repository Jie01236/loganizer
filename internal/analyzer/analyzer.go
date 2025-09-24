package analyzer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"sync"
	"time"

	"loganalyzer/internal/config"
)

const (
	StatusOK = "OK"
	StatusFailed = "FAILED"
)

type LogAnalysisResult struct {
	LogID        string `json:"log_id"`
	FilePath     string `json:"file_path"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	ErrorDetails string `json:"error_details,omitempty"`
}

var ErrFileAccess = errors.New("file access error")

// Erreurs Personnalisées：FileAccessError offre le context d'un log file qui n'est pas accessible.
type FileAccessError struct {
	Path string
	Err  error
}

func (e *FileAccessError) Error() string {
	return fmt.Sprintf("accès impossible au fichier %q: %v", e.Path, e.Err)
}

func (e *FileAccessError) Unwrap() error {
	return e.Err
}

func (e *FileAccessError) Is(target error) bool {
	return target == ErrFileAccess
}

// Concurrence: function Analyze lance des goroutines pour chaque entrée et collecte les résultats.
func Analyze(ctx context.Context, entries []config.LogConfig) []LogAnalysisResult {
	results := make([]LogAnalysisResult, len(entries))
	if len(entries) == 0 {
		return results
	}

	type indexedResult struct {
		index  int
		result LogAnalysisResult
	}

	var wg sync.WaitGroup
	wg.Add(len(entries))

	out := make(chan indexedResult, len(entries))

	for idx, entry := range entries {
		idx := idx
		entry := entry

		go func() {
			defer wg.Done()

			res := LogAnalysisResult{
				LogID:    entry.ID,
				FilePath: entry.Path,
			}

			// Vérifie l'accessibilité du fichier
			info, err := os.Stat(entry.Path)
			if err != nil {
				err = &FileAccessError{Path: entry.Path, Err: err}
				res.Status = StatusFailed
				msg, details := buildFailureMessages(err)
				res.Message = msg
				res.ErrorDetails = details
				out <- indexedResult{index: idx, result: res}
				return
			}
			if info.IsDir() {
				err = &FileAccessError{Path: entry.Path, Err: fmt.Errorf("path is a directory")}
				res.Status = StatusFailed
				msg, details := buildFailureMessages(err)
				res.Message = msg
				res.ErrorDetails = details
				out <- indexedResult{index: idx, result: res}
				return
			}

			// Open: vérifie la lisibilité réelle
			f, err := os.Open(entry.Path)
			if err != nil {
				err = &FileAccessError{Path: entry.Path, Err: err}
				res.Status = StatusFailed
				msg, details := buildFailureMessages(err)
				res.Message = msg
				res.ErrorDetails = details
				out <- indexedResult{index: idx, result: res}
				return
			}
			_ = f.Close()

			if err := simulateWork(ctx, idx); err != nil {
				res.Status = StatusFailed
				res.Message = "Analyse annulée."
				res.ErrorDetails = err.Error()
				out <- indexedResult{index: idx, result: res}
				return
			}

			res.Status = StatusOK
			res.Message = "Analyse terminée avec succès."
			out <- indexedResult{index: idx, result: res}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	for item := range out {
		results[item.index] = item.result
	}

	return results
}

// FilterByStatus renvoie seulement les résultats correspondant au statut demandé.
func FilterByStatus(results []LogAnalysisResult, status string) []LogAnalysisResult {
	if status == "" {
		return results
	}

	filtered := make([]LogAnalysisResult, 0, len(results))
	for _, res := range results {
		if res.Status == status {
			filtered = append(filtered, res)
		}
	}

	return filtered
}

func buildFailureMessages(err error) (message string, details string) {
	var accessErr *FileAccessError
	if errors.As(err, &accessErr) {
		switch {
		case errors.Is(accessErr.Err, fs.ErrNotExist):
			message = "Fichier introuvable."
		case errors.Is(accessErr.Err, os.ErrPermission):
			message = "Accès refusé."
		default:
			message = "Impossible de lire le fichier."
		}
		details = accessErr.Err.Error()
		return
	}

	message = "Analyse interrompue."
	details = err.Error()
	return
}

func simulateWork(ctx context.Context, seed int) error {
	duration := randomDuration(seed)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func randomDuration(seed int) time.Duration {
	source := rand.NewSource(time.Now().UnixNano() + int64(seed))
	r := rand.New(source)
	milliseconds := r.Intn(151) + 50
	return time.Duration(milliseconds) * time.Millisecond
}
