package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hidetzu/prism/internal/config"
	"github.com/hidetzu/prism/internal/domain"
	ghprovider "github.com/hidetzu/prism/internal/provider/github"
	"github.com/hidetzu/prism/internal/usecase"
)

const version = "0.1.0-dev"

// Exit codes as defined in docs/spec.md.
const (
	ExitSuccess       = 0
	ExitGeneralError  = 1
	ExitInvalidArgs   = 2
	ExitProviderError = 3
	ExitAnalysisError = 4
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidArgs):
		return ExitInvalidArgs
	case errors.Is(err, domain.ErrProvider):
		return ExitProviderError
	case errors.Is(err, domain.ErrAnalysis):
		return ExitAnalysisError
	default:
		return ExitGeneralError
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "prism",
		Short:         "Review Context Compiler — transform PRs into AI-review-ready input",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		analyzeCmd(),
		promptCmd(),
		fetchCmd(),
	)

	return cmd
}

func analyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <PR_URL>",
		Short: "Fetch a PR, run analysis, and output structured results",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAnalyze,
	}

	cmd.Flags().String("format", "json", "Output format (json|markdown|text)")
	cmd.Flags().String("config", "", "Config file path")

	return cmd
}

func promptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt <PR_URL>",
		Short: "Generate a review prompt for AI consumption",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runPrompt,
	}

	cmd.Flags().String("mode", "light", "Prompt mode (light|detailed|cross)")
	cmd.Flags().String("format", "text", "Output format (text|markdown|json)")
	cmd.Flags().String("lang", "en", "Prompt language (en|ja)")
	cmd.Flags().String("template", "", "Custom template path")
	cmd.Flags().String("config", "", "Config file path")

	return cmd
}

func fetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <PR_URL>",
		Short: "Fetch raw PR data for debugging (no analysis)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runFetch,
	}

	cmd.Flags().String("format", "json", "Output format (json|text)")
	cmd.Flags().String("config", "", "Config file path")

	return cmd
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: PR URL is required", domain.ErrInvalidArgs)
	}

	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	p := newProvider(cfg)
	ref, err := p.Parse(args[0])
	if err != nil {
		return fmt.Errorf("%w: invalid PR URL: %v", domain.ErrInvalidArgs, err)
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "" {
		format = cfg.DefaultFormat
	}

	return usecase.Analyze(cmd.Context(), p, ref, usecase.AnalyzeOptions{
		Format: format,
	}, os.Stdout)
}

func runPrompt(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: PR URL is required", domain.ErrInvalidArgs)
	}

	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	p := newProvider(cfg)
	ref, err := p.Parse(args[0])
	if err != nil {
		return fmt.Errorf("%w: invalid PR URL: %v", domain.ErrInvalidArgs, err)
	}

	mode, _ := cmd.Flags().GetString("mode")
	format, _ := cmd.Flags().GetString("format")
	lang, _ := cmd.Flags().GetString("lang")
	tmpl, _ := cmd.Flags().GetString("template")

	if mode == "" && cfg.DefaultMode != "" {
		mode = cfg.DefaultMode
	}
	if lang == "" && cfg.DefaultLang != "" {
		lang = cfg.DefaultLang
	}

	return usecase.Prompt(cmd.Context(), p, ref, usecase.PromptOptions{
		Mode:     mode,
		Format:   format,
		Lang:     lang,
		Template: tmpl,
	}, os.Stdout)
}

func runFetch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: PR URL is required", domain.ErrInvalidArgs)
	}

	configPath, _ := cmd.Flags().GetString("config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	p := newProvider(cfg)
	ref, err := p.Parse(args[0])
	if err != nil {
		return fmt.Errorf("%w: invalid PR URL: %v", domain.ErrInvalidArgs, err)
	}

	format, _ := cmd.Flags().GetString("format")

	return usecase.Fetch(cmd.Context(), p, ref, usecase.FetchOptions{
		Format: format,
	}, os.Stdout)
}

func newProvider(cfg config.Config) *ghprovider.Provider {
	token := cfg.GitHubToken
	return ghprovider.NewProvider(token)
}
