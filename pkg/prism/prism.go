// Package prism is the public API for embedding prism as a library.
//
// prism decomposes pull requests into structured, AI-review-ready context.
// This package is the stable entry point for programmatic use by prism-api,
// editor plugins, CI integrations, and other third-party consumers.
//
// Typical usage:
//
//	result, err := prism.Analyze(ctx, prism.AnalyzeOptions{
//	    PRURL:       "https://github.com/owner/repo/pull/123",
//	    GitHubToken: os.Getenv("GITHUB_TOKEN"),
//	})
//
// Compatibility: within a major version, Analyze and Prompt signatures do not
// change, and existing fields on exported types do not change type or meaning.
// New optional fields may be added in minor versions. See ADR-0002 for details.
package prism

import (
	"context"
	"errors"
	"fmt"

	"github.com/hidetzu/prism/internal/analyzer"
	"github.com/hidetzu/prism/internal/classifier"
	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/prompt"
	"github.com/hidetzu/prism/internal/provider"
)

// AnalyzeOptions configures a prism Analyze or Prompt call.
type AnalyzeOptions struct {
	// Provider is the provider name (e.g. "github", "codecommit").
	// If empty, the provider is auto-detected from PRURL.
	Provider string

	// PRURL is the pull request URL.
	PRURL string

	// GitHubToken is the authentication token for the GitHub provider.
	// Not used for plugin providers that manage their own authentication.
	GitHubToken string

	// IncludePatches controls whether ChangedFile.Patch is populated.
	// Default: false (patches excluded to keep responses lightweight).
	IncludePatches bool

	// Mode is the prompt mode: "light", "detailed", or "cross".
	// Only used by Prompt. Defaults to "light" if empty.
	Mode string

	// Language is the prompt language: "en" or "ja".
	// Only used by Prompt. Defaults to "en" if empty.
	Language string
}

// PRInfo is the minimal pull request metadata returned with every Result.
type PRInfo struct {
	Provider     string `json:"provider"`
	Repository   string `json:"repository"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	SourceBranch string `json:"source_branch,omitempty"`
	TargetBranch string `json:"target_branch,omitempty"`
	URL          string `json:"url"`
}

// ChangedFile represents a file changed in the pull request.
type ChangedFile struct {
	Path        string `json:"path"`
	Status      string `json:"status"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Language    string `json:"language,omitempty"`
	IsTest      bool   `json:"is_test,omitempty"`
	IsConfig    bool   `json:"is_config,omitempty"`
	IsGenerated bool   `json:"is_generated,omitempty"`
	// Patch is populated only when AnalyzeOptions.IncludePatches is true.
	Patch string `json:"patch,omitempty"`
}

// AnalysisResult holds the structured analysis output.
type AnalysisResult struct {
	ChangeType    string   `json:"change_type"`
	RiskLevel     string   `json:"risk_level"`
	AffectedAreas []string `json:"affected_areas,omitempty"`
	ReviewAxes    []string `json:"review_axes,omitempty"`
	RelatedFiles  []string `json:"related_files,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
	Summary       string   `json:"summary,omitempty"`
}

// Result is the complete output of Analyze.
// It mirrors the CLI JSON schema so that CLI and library callers see the same shape.
type Result struct {
	PR       PRInfo         `json:"pull_request"`
	Analysis AnalysisResult `json:"analysis"`
	Files    []ChangedFile  `json:"changed_files,omitempty"`
}

// Sentinel errors for client-side branching (e.g. HTTP status mapping).
// Callers can use errors.Is to check for these categories.
var (
	// ErrInvalidInput indicates malformed or missing input (bad URL, empty options, etc.).
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnsupportedProvider indicates the provider is not recognized or the plugin is not installed.
	ErrUnsupportedProvider = errors.New("unsupported provider")

	// ErrAuthRequired indicates authentication is missing or failed.
	// Currently reserved for future use; provider implementations will map
	// 401/403 responses to this error. See https://github.com/hidetzu/prism/issues
	// for tracking.
	ErrAuthRequired = errors.New("authentication required")

	// ErrUpstreamFailure indicates the upstream provider API failed (network, 5xx, timeout).
	ErrUpstreamFailure = errors.New("upstream failure")
)

// Analyze fetches a pull request and returns structured analysis.
func Analyze(ctx context.Context, opts AnalyzeOptions) (Result, error) {
	if opts.PRURL == "" {
		return Result{}, fmt.Errorf("%w: PRURL is required", ErrInvalidInput)
	}

	p, ref, err := resolveProvider(opts)
	if err != nil {
		return Result{}, err
	}

	pr, err := p.FetchPullRequest(ctx, ref)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUpstreamFailure, err)
	}

	changeType := classifier.Classify(pr)
	analysis := analyzer.Analyze(pr, changeType)

	return buildResult(ref.Provider, pr, analysis, opts.IncludePatches), nil
}

// Prompt fetches a pull request and returns a review prompt string.
// The returned string contains the system prompt and user prompt separated by
// "\n\n---\n\n", matching the default CLI text format.
func Prompt(ctx context.Context, opts AnalyzeOptions) (string, error) {
	if opts.PRURL == "" {
		return "", fmt.Errorf("%w: PRURL is required", ErrInvalidInput)
	}

	mode := opts.Mode
	if mode == "" {
		mode = "light"
	}
	if mode != "light" && mode != "detailed" && mode != "cross" {
		return "", fmt.Errorf("%w: invalid mode %q (must be light, detailed, or cross)", ErrInvalidInput, mode)
	}

	lang := opts.Language
	if lang == "" {
		lang = "en"
	}
	if lang != "en" && lang != "ja" {
		return "", fmt.Errorf("%w: invalid language %q (must be en or ja)", ErrInvalidInput, lang)
	}

	p, ref, err := resolveProvider(opts)
	if err != nil {
		return "", err
	}

	pr, err := p.FetchPullRequest(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpstreamFailure, err)
	}

	changeType := classifier.Classify(pr)
	analysis := analyzer.Analyze(pr, changeType)

	bundle := prompt.Render(domain.PromptMode(mode), pr, analysis, lang)
	return fmt.Sprintf("%s\n\n---\n\n%s", bundle.SystemPrompt, bundle.UserPrompt), nil
}

// resolveProvider builds the provider and PRRef from AnalyzeOptions.
func resolveProvider(opts AnalyzeOptions) (provider.Provider, domain.PRRef, error) {
	reg := provider.NewRegistry(opts.GitHubToken)
	p, err := reg.Resolve(opts.Provider, opts.PRURL)
	if err != nil {
		// Registry errors are either unknown host (invalid) or plugin-not-found (unsupported).
		if opts.Provider != "" {
			return nil, domain.PRRef{}, fmt.Errorf("%w: %v", ErrUnsupportedProvider, err)
		}
		return nil, domain.PRRef{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	ref, err := p.Parse(opts.PRURL)
	if err != nil {
		return nil, domain.PRRef{}, fmt.Errorf("%w: invalid PR URL: %v", ErrInvalidInput, err)
	}

	return p, ref, nil
}

// buildResult converts internal domain models into the public Result shape.
func buildResult(providerName string, pr domain.PullRequest, analysis domain.AnalysisResult, includePatches bool) Result {
	files := make([]ChangedFile, 0, len(pr.ChangedFiles))
	for _, f := range pr.ChangedFiles {
		cf := ChangedFile{
			Path:        f.Path,
			Status:      string(f.Status),
			Additions:   f.Additions,
			Deletions:   f.Deletions,
			Language:    f.Language,
			IsTest:      f.IsTest,
			IsConfig:    f.IsConfig,
			IsGenerated: f.IsGenerated,
		}
		if includePatches {
			cf.Patch = f.Patch
		}
		files = append(files, cf)
	}

	axes := make([]string, 0, len(analysis.ReviewAxes))
	for _, a := range analysis.ReviewAxes {
		axes = append(axes, string(a))
	}

	return Result{
		PR: PRInfo{
			Provider:     providerName,
			Repository:   pr.Repository,
			ID:           pr.ID,
			Title:        pr.Title,
			Author:       pr.Author,
			SourceBranch: pr.SourceBranch,
			TargetBranch: pr.TargetBranch,
			URL:          buildPRURL(providerName, pr.Repository, pr.ID),
		},
		Analysis: AnalysisResult{
			ChangeType:    string(analysis.ChangeType),
			RiskLevel:     string(analysis.RiskLevel),
			AffectedAreas: analysis.AffectedAreas,
			ReviewAxes:    axes,
			RelatedFiles:  analysis.RelatedFiles,
			Warnings:      analysis.Warnings,
			Summary:       analysis.Summary,
		},
		Files: files,
	}
}

// buildPRURL reconstructs the canonical PR URL from provider, repository, and ID.
// Returns an empty string if the provider is unknown.
func buildPRURL(providerName, repository, id string) string {
	switch providerName {
	case "github":
		if repository == "" || id == "" {
			return ""
		}
		return fmt.Sprintf("https://github.com/%s/pull/%s", repository, id)
	default:
		return ""
	}
}
