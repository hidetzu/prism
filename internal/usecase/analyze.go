package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/hidetzu/prism/internal/analyzer"
	"github.com/hidetzu/prism/internal/classifier"
	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/formatter"
	"github.com/hidetzu/prism/internal/provider"
)

// AnalyzeOptions holds options for the Analyze use case.
type AnalyzeOptions struct {
	Format string // "json", "markdown", "text"
}

// Analyze fetches a PR, classifies and analyzes it, and writes formatted output.
func Analyze(ctx context.Context, p provider.Provider, ref domain.PRRef, opts AnalyzeOptions, w io.Writer) error {
	if err := ValidateAnalyzeOptions(opts); err != nil {
		return err
	}

	pr, err := p.FetchPullRequest(ctx, ref)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrProvider, err)
	}

	changeType := classifier.Classify(pr)
	result := analyzer.Analyze(pr, changeType)

	switch opts.Format {
	case "markdown":
		return formatter.FormatMarkdown(w, ref.Provider, pr, result)
	case "text":
		return formatter.FormatText(w, ref.Provider, pr, result)
	default:
		return formatter.FormatJSON(w, ref.Provider, pr, result)
	}
}
