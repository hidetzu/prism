package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/hidetzu/prism/pkg/prism"
)

// FormatText writes a prism.Result as plain text to w.
func FormatText(w io.Writer, result prism.Result) error {
	write := func(format string, args ...interface{}) error {
		_, err := fmt.Fprintf(w, format, args...)
		return err
	}

	if err := write("%s\n", result.PR.Title); err != nil {
		return err
	}
	if err := write("%s\n\n", strings.Repeat("=", len(result.PR.Title))); err != nil {
		return err
	}

	if err := write("Repository:  %s\n", result.PR.Repository); err != nil {
		return err
	}
	if err := write("PR:          #%s\n", result.PR.ID); err != nil {
		return err
	}
	if err := write("Author:      %s\n", result.PR.Author); err != nil {
		return err
	}
	if err := write("Branch:      %s -> %s\n", result.PR.SourceBranch, result.PR.TargetBranch); err != nil {
		return err
	}
	if err := write("Provider:    %s\n", result.PR.Provider); err != nil {
		return err
	}
	if result.PR.URL != "" {
		if err := write("URL:         %s\n", result.PR.URL); err != nil {
			return err
		}
	}
	if err := write("\n"); err != nil {
		return err
	}

	if err := write("Change Type: %s\n", result.Analysis.ChangeType); err != nil {
		return err
	}
	if err := write("Risk Level:  %s\n", result.Analysis.RiskLevel); err != nil {
		return err
	}
	if result.Analysis.Summary != "" {
		if err := write("Summary:     %s\n", result.Analysis.Summary); err != nil {
			return err
		}
	}
	if err := write("\n"); err != nil {
		return err
	}

	if len(result.Analysis.ReviewAxes) > 0 {
		if err := write("Review Axes:\n"); err != nil {
			return err
		}
		for _, axis := range result.Analysis.ReviewAxes {
			if err := write("  - %s\n", axis); err != nil {
				return err
			}
		}
		if err := write("\n"); err != nil {
			return err
		}
	}

	if len(result.Analysis.AffectedAreas) > 0 {
		if err := write("Affected Areas:\n"); err != nil {
			return err
		}
		for _, area := range result.Analysis.AffectedAreas {
			if err := write("  - %s\n", area); err != nil {
				return err
			}
		}
		if err := write("\n"); err != nil {
			return err
		}
	}

	if len(result.Analysis.Warnings) > 0 {
		if err := write("Warnings:\n"); err != nil {
			return err
		}
		for _, warn := range result.Analysis.Warnings {
			if err := write("  - %s\n", warn); err != nil {
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
	for _, f := range result.Files {
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
