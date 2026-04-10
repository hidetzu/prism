package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/hidetzu/prism/internal/domain"
)

// FormatText writes the analysis result as plain text to w.
func FormatText(w io.Writer, providerName string, pr domain.PullRequest, result domain.AnalysisResult) error {
	write := func(format string, args ...interface{}) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}

	if err := write("%s\n", pr.Title); err != nil {
		return err
	}
	if err := write("%s\n\n", strings.Repeat("=", len(pr.Title))); err != nil {
		return err
	}

	if err := write("Repository:  %s\n", pr.Repository); err != nil {
		return err
	}
	if err := write("PR:          #%s\n", pr.ID); err != nil {
		return err
	}
	if err := write("Author:      %s\n", pr.Author); err != nil {
		return err
	}
	if err := write("Branch:      %s -> %s\n", pr.SourceBranch, pr.TargetBranch); err != nil {
		return err
	}
	if err := write("Provider:    %s\n\n", providerName); err != nil {
		return err
	}

	if err := write("Change Type: %s\n", result.ChangeType); err != nil {
		return err
	}
	if err := write("Risk Level:  %s\n", result.RiskLevel); err != nil {
		return err
	}
	if result.Summary != "" {
		if err := write("Summary:     %s\n", result.Summary); err != nil {
			return err
		}
	}
	if err := write("\n"); err != nil {
		return err
	}

	if len(result.ReviewAxes) > 0 {
		if err := write("Review Axes:\n"); err != nil {
			return err
		}
		for _, axis := range result.ReviewAxes {
			if err := write("  - %s\n", axis); err != nil {
				return err
			}
		}
		if err := write("\n"); err != nil {
			return err
		}
	}

	if len(result.AffectedAreas) > 0 {
		if err := write("Affected Areas:\n"); err != nil {
			return err
		}
		for _, area := range result.AffectedAreas {
			if err := write("  - %s\n", area); err != nil {
				return err
			}
		}
		if err := write("\n"); err != nil {
			return err
		}
	}

	if len(result.Warnings) > 0 {
		if err := write("Warnings:\n"); err != nil {
			return err
		}
		for _, w := range result.Warnings {
			if err := write("  - %s\n", w); err != nil {
				return err
			}
		}
		if err := write("\n"); err != nil {
			return err
		}
	}

	if err := write("Changed Files:\n"); err != nil {
		return err
	}
	for _, f := range pr.ChangedFiles {
		flags := fileFlags(f)
		if flags != "" {
			flags = " " + flags
		}
		if err := write("  %s (%s, +%d/-%d, %s)%s\n", f.Path, f.Status, f.Additions, f.Deletions, f.Language, flags); err != nil {
			return err
		}
	}

	return nil
}
