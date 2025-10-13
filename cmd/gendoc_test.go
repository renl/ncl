package cmd

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/renl/ncl/internal/config"
)

func newTestGendocCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "gendoc"}
	cmd.Flags().StringVar(&gendocAPIKey, "api-key", "", "")
	cmd.Flags().StringVar(&gendocBaseURL, "base-url", "", "")
	cmd.Flags().StringVar(&gendocModel, "model", "gpt-4o-mini", "")
	return cmd
}

func TestApplyGendocConfigValues_AppliesDefaults(t *testing.T) {
	gendocAPIKey = ""
	gendocBaseURL = ""
	gendocModel = "gpt-4o-mini"

	cmd := newTestGendocCommand()
	cfg := &config.Config{
		SourcePath: "/tmp/.nclrc",
		Gendoc: config.GendocSection{
			APIKey:  "CONFIG_KEY",
			BaseURL: "https://example.com",
			Model:   "config-model",
		},
	}

	applyGendocConfigValues(cmd, cfg)

	if gendocAPIKey != "CONFIG_KEY" {
		t.Fatalf("expected api key from config, got %q", gendocAPIKey)
	}
	if gendocBaseURL != "https://example.com" {
		t.Fatalf("expected base url from config, got %q", gendocBaseURL)
	}
	if gendocModel != "config-model" {
		t.Fatalf("expected model from config, got %q", gendocModel)
	}
}

func TestApplyGendocConfigValues_FlagOverrides(t *testing.T) {
	gendocAPIKey = "flag-key"
	gendocBaseURL = "flag-base"
	gendocModel = "flag-model"

	cmd := newTestGendocCommand()
	if err := cmd.Flags().Set("api-key", "flag-key"); err != nil {
		t.Fatalf("Set api-key: %v", err)
	}
	if err := cmd.Flags().Set("base-url", "flag-base"); err != nil {
		t.Fatalf("Set base-url: %v", err)
	}
	if err := cmd.Flags().Set("model", "flag-model"); err != nil {
		t.Fatalf("Set model: %v", err)
	}

	cfg := &config.Config{
		SourcePath: "/tmp/.nclrc",
		Gendoc: config.GendocSection{
			APIKey:  "CONFIG_KEY",
			BaseURL: "https://example.com",
			Model:   "config-model",
		},
	}

	applyGendocConfigValues(cmd, cfg)

	if gendocAPIKey != "flag-key" {
		t.Fatalf("expected api key to remain flag value, got %q", gendocAPIKey)
	}
	if gendocBaseURL != "flag-base" {
		t.Fatalf("expected base url to remain flag value, got %q", gendocBaseURL)
	}
	if gendocModel != "flag-model" {
		t.Fatalf("expected model to remain flag value, got %q", gendocModel)
	}
}
