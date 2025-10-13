/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/renl/ncl/internal/cmdutil"
	"github.com/renl/ncl/internal/config"
	"github.com/renl/ncl/internal/techlead"
)

var (
	gendocModel       string
	gendocProvider    string
	gendocOutPath     string
	gendocTemperature float64
	gendocAPIKey      string
	gendocBaseURL     string
)

// gendocCmd represents the gendoc command
var gendocCmd = &cobra.Command{
	Use:   "gendoc",
	Short: "Co-author a Tech Lead plan with an LLM guided session",
	Long: `Start an interactive CLI workshop that gathers project context and partners
with an LLM to produce a Tech Lead planning document. After the first draft,
iterate by sending refinement instructions, saving checkpoints, and exporting
the final Markdown.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := applyGendocConfig(cmd); err != nil {
			return err
		}

		if err := validateTemperature(gendocTemperature); err != nil {
			return err
		}

		model, callOpts, err := buildModel(gendocProvider, gendocModel, gendocAPIKey, gendocBaseURL)
		if err != nil {
			return err
		}

		cmdutil.VPrintf(cmd, "using provider=%s model=%s temperature=%.2f", gendocProvider, gendocModel, gendocTemperature)

		session := techlead.NewSession(model, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), techlead.SessionConfig{
			DefaultOutputPath: strings.TrimSpace(gendocOutPath),
			SystemPrompt:      "",
			CallOptions:       append(callOpts, llms.WithTemperature(gendocTemperature)),
		})

		return session.Run(cmd.Context())
	},
}

func applyGendocConfig(cmd *cobra.Command) error {
	cfg, err := config.LoadRC()
	if err != nil {
		return err
	}
	applyGendocConfigValues(cmd, cfg)
	return nil
}

func applyGendocConfigValues(cmd *cobra.Command, cfg *config.Config) {
	if cfg == nil {
		return
	}

	if cfg.Gendoc.APIKey != "" && !cmd.Flags().Changed("api-key") {
		gendocAPIKey = strings.TrimSpace(cfg.Gendoc.APIKey)
	}
	if cfg.Gendoc.BaseURL != "" && !cmd.Flags().Changed("base-url") {
		gendocBaseURL = strings.TrimSpace(cfg.Gendoc.BaseURL)
	}
	if cfg.Gendoc.Model != "" && !cmd.Flags().Changed("model") {
		gendocModel = strings.TrimSpace(cfg.Gendoc.Model)
	}

	if cfg.SourcePath != "" {
		cmdutil.VPrintf(cmd, "loaded config from %s", cfg.SourcePath)
	}
}

func init() {
	techleadCmd.AddCommand(gendocCmd)

	gendocCmd.Flags().StringVarP(&gendocOutPath, "out", "o", "", "File to write the final Markdown output on exit")
	gendocCmd.Flags().StringVar(&gendocModel, "model", "gpt-4o-mini", "LLM model name (provider specific)")
	gendocCmd.Flags().StringVar(&gendocProvider, "provider", "openai", "LLM provider to use (currently only 'openai' is supported)")
	gendocCmd.Flags().Float64Var(&gendocTemperature, "temperature", 0.2, "LLM generation temperature (0.0 – 2.0)")
	gendocCmd.Flags().StringVar(&gendocAPIKey, "api-key", "", "Override API key (defaults to provider-specific env vars)")
	gendocCmd.Flags().StringVar(&gendocBaseURL, "base-url", "", "Override base URL for the provider API (advanced)")
}

func buildModel(provider, model, apiKey, baseURL string) (llms.Model, []llms.CallOption, error) {
	switch strings.ToLower(provider) {
	case "openai":
		return buildOpenAI(model, apiKey, baseURL)
	default:
		return nil, nil, fmt.Errorf("unsupported provider %q", provider)
	}
}

func buildOpenAI(model, apiKey, baseURL string) (llms.Model, []llms.CallOption, error) {
	var opts []openai.Option
	if strings.TrimSpace(model) != "" {
		opts = append(opts, openai.WithModel(model))
	}
	if strings.TrimSpace(apiKey) != "" {
		opts = append(opts, openai.WithToken(strings.TrimSpace(apiKey)))
	}
	if strings.TrimSpace(baseURL) != "" {
		opts = append(opts, openai.WithBaseURL(strings.TrimSpace(baseURL)))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		if errors.Is(err, openai.ErrMissingToken) {
			return nil, nil, fmt.Errorf("%w (set OPENAI_API_KEY or pass --api-key)", err)
		}
		return nil, nil, err
	}
	return llm, nil, nil
}

func validateTemperature(value float64) error {
	if math.IsNaN(value) || value < 0 || value > 2 {
		return errors.New("temperature must be between 0 and 2")
	}
	return nil
}
