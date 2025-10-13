package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRCFromPaths_FirstMatchWins(t *testing.T) {
	dir := t.TempDir()

	other := filepath.Join(dir, "other.yaml")
	if err := os.WriteFile(other, []byte("gendoc:\n  api_key: SECOND"), 0o644); err != nil {
		t.Fatalf("write other: %v", err)
	}

	target := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(target, []byte("gendoc:\n  api_key: FIRST\n  model: model-x"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	cfg, err := LoadRCFromPaths([]string{"missing.yaml", target, other})
	if err != nil {
		t.Fatalf("LoadRCFromPaths: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected config to be loaded")
	}
	if cfg.Gendoc.APIKey != "FIRST" {
		t.Fatalf("unexpected api key: %s", cfg.Gendoc.APIKey)
	}
	if cfg.Gendoc.Model != "model-x" {
		t.Fatalf("unexpected model: %s", cfg.Gendoc.Model)
	}
	if cfg.SourcePath != filepath.Clean(target) {
		t.Fatalf("unexpected source path: %s", cfg.SourcePath)
	}
}

func TestLoadRCFromPaths_IgnoresMissingAndErrors(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(bad, []byte("not: [valid"), 0o644); err != nil {
		t.Fatalf("write bad: %v", err)
	}

	if _, err := LoadRCFromPaths([]string{bad}); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestLoadRC_NoFile(t *testing.T) {
	cfg, err := LoadRCFromPaths([]string{"nonexistent"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config when no file exists")
	}
}
