package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents settings loaded from a .nclrc file.
type Config struct {
	SourcePath string        `yaml:"-"`
	Gendoc     GendocSection `yaml:"gendoc"`
}

// GendocSection encapsulates configuration for the techlead gendoc command.
type GendocSection struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

// LoadRC searches for a .nclrc file and returns the parsed configuration.
//
// Search order:
//  1. ./.nclrc
//  2. ~/.nclrc
//  3. ~/.ncl/.nclrc
//
// If no file is found, (nil, "", nil) is returned.
func LoadRC() (*Config, error) {
	paths := candidatePaths()
	return loadFromPaths(paths)
}

// LoadRCFromPaths attempts to load the first readable config from the provided paths.
// Intended for tests.
func LoadRCFromPaths(paths []string) (*Config, error) {
	if len(paths) == 0 {
		return nil, errors.New("paths cannot be empty")
	}
	return loadFromPaths(paths)
}

func loadFromPaths(paths []string) (*Config, error) {
	for _, p := range paths {
		expanded := filepath.Clean(p)
		data, err := os.ReadFile(expanded)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read config %s: %w", expanded, err)
		}

		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", expanded, err)
		}
		cfg.SourcePath = expanded
		return &cfg, nil
	}
	return nil, nil
}

func candidatePaths() []string {
	cwd, err := os.Getwd()
	var paths []string
	if err == nil {
		paths = append(paths, filepath.Join(cwd, ".nclrc"))
	} else {
		paths = append(paths, ".nclrc")
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		paths = append(paths, filepath.Join(home, ".nclrc"))
		paths = append(paths, filepath.Join(home, ".ncl", ".nclrc"))
	}

	return paths
}
