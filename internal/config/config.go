package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type LogConfig struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
}

var ErrConfigParse = errors.New("config parse error")

// ParseError pour unwrap() pour montrer les details de configuration.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("impossible d'analyser le fichier de configuration %q: %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func (e *ParseError) Is(target error) bool {
	return target == ErrConfigParse
}

// Load lit the log configuration.
func Load(path string) ([]LogConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lecture du fichier de configuration %q: %w", path, err)
	}

	var entries []LogConfig
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	for idx, entry := range entries {
		if strings.TrimSpace(entry.ID) == "" {
			return nil, &ParseError{Path: path, Err: fmt.Errorf("entrée %d: champ id manquant", idx)}
		}
		if strings.TrimSpace(entry.Path) == "" {
			return nil, &ParseError{Path: path, Err: fmt.Errorf("entrée %d: champ path manquant", idx)}
		}
		if strings.TrimSpace(entry.Type) == "" {
			return nil, &ParseError{Path: path, Err: fmt.Errorf("entrée %d: champ type manquant", idx)}
		}
	}

	return entries, nil
}
