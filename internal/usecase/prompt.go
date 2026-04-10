package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hidetzu/prism/internal/analyzer"
	"github.com/hidetzu/prism/internal/classifier"
	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/prompt"
	"github.com/hidetzu/prism/internal/provider"
)

// PromptOptions holds options for the Prompt use case.
type PromptOptions struct {
	Mode     string // "light", "detailed", "cross"
	Format   string // "text", "markdown", "json"
	Lang     string // "en", "ja"
	Template string // custom template file path
}

// Prompt fetches a PR, analyzes it, and generates a review prompt.
func Prompt(ctx context.Context, p provider.Provider, ref domain.PRRef, opts PromptOptions, w io.Writer) error {
	if err := ValidatePromptOptions(opts); err != nil {
		return err
	}

	pr, err := p.FetchPullRequest(ctx, ref)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrProvider, err)
	}

	changeType := classifier.Classify(pr)
	result := analyzer.Analyze(pr, changeType)

	mode := domain.PromptMode(opts.Mode)
	lang := opts.Lang
	if lang == "" {
		lang = "en"
	}

	// Custom template overrides built-in rendering.
	if opts.Template != "" {
		bundle, err := prompt.RenderFromTemplate(opts.Template, mode, pr, result, lang)
		if err != nil {
			return fmt.Errorf("render template: %w", err)
		}
		return writePromptBundle(w, opts.Format, bundle)
	}

	bundle := prompt.Render(mode, pr, result, lang)

	return writePromptBundle(w, opts.Format, bundle)
}

func writePromptBundle(w io.Writer, format string, bundle domain.PromptBundle) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(bundle)
	case "markdown":
		_, err := fmt.Fprintf(w, "## System Prompt\n\n%s\n\n---\n\n%s", bundle.SystemPrompt, bundle.UserPrompt)
		return err
	default:
		_, err := fmt.Fprintf(w, "%s\n\n---\n\n%s", bundle.SystemPrompt, bundle.UserPrompt)
		return err
	}
}
