package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider"
)

// FetchOptions holds options for the Fetch use case.
type FetchOptions struct {
	Format string // "json", "text"
}

// Fetch retrieves raw PR data without analysis.
func Fetch(ctx context.Context, p provider.Provider, ref domain.PRRef, opts FetchOptions, w io.Writer) error {
	if err := ValidateFetchOptions(opts); err != nil {
		return err
	}

	pr, err := p.FetchPullRequest(ctx, ref)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrProvider, err)
	}

	switch opts.Format {
	case "text":
		return writeTextPR(w, pr)
	default:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(pr)
	}
}

func writeTextPR(w io.Writer, pr domain.PullRequest) error {
	write := func(format string, args ...interface{}) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}

	if err := write("Repository: %s\n", pr.Repository); err != nil {
		return err
	}
	if err := write("PR: #%s\n", pr.ID); err != nil {
		return err
	}
	if err := write("Title: %s\n", pr.Title); err != nil {
		return err
	}
	if err := write("Author: %s\n", pr.Author); err != nil {
		return err
	}
	if err := write("Branch: %s -> %s\n", pr.SourceBranch, pr.TargetBranch); err != nil {
		return err
	}
	if pr.Description != "" {
		if err := write("\n%s\n", pr.Description); err != nil {
			return err
		}
	}
	if err := write("\nChanged Files (%d):\n", len(pr.ChangedFiles)); err != nil {
		return err
	}
	for _, f := range pr.ChangedFiles {
		if err := write("  %s (%s, +%d/-%d)\n", f.Path, f.Status, f.Additions, f.Deletions); err != nil {
			return err
		}
	}
	return nil
}
