package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/hidetzu/prism/internal/domain"
)

// FormatMarkdown writes the analysis result as human-readable Markdown to w.
func FormatMarkdown(w io.Writer, providerName string, pr domain.PullRequest, result domain.AnalysisResult) error {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "# %s\n\n", pr.Title)

	// Metadata
	b.WriteString("## Pull Request\n\n")
	fmt.Fprintf(&b, "| Field | Value |\n")
	fmt.Fprintf(&b, "|-------|-------|\n")
	fmt.Fprintf(&b, "| Repository | %s |\n", pr.Repository)
	fmt.Fprintf(&b, "| PR | #%s |\n", pr.ID)
	fmt.Fprintf(&b, "| Author | %s |\n", pr.Author)
	fmt.Fprintf(&b, "| Branch | %s -> %s |\n", pr.SourceBranch, pr.TargetBranch)
	fmt.Fprintf(&b, "| Provider | %s |\n", providerName)
	b.WriteString("\n")

	if pr.Description != "" {
		b.WriteString("### Description\n\n")
		b.WriteString(pr.Description)
		b.WriteString("\n\n")
	}

	// Analysis
	b.WriteString("## Analysis\n\n")
	fmt.Fprintf(&b, "- **Change Type:** %s\n", result.ChangeType)
	fmt.Fprintf(&b, "- **Risk Level:** %s\n", result.RiskLevel)
	fmt.Fprintf(&b, "- **Summary:** %s\n", result.Summary)
	b.WriteString("\n")

	// Review Axes
	if len(result.ReviewAxes) > 0 {
		b.WriteString("### Review Axes\n\n")
		for _, axis := range result.ReviewAxes {
			fmt.Fprintf(&b, "- %s\n", axis)
		}
		b.WriteString("\n")
	}

	// Affected Areas
	if len(result.AffectedAreas) > 0 {
		b.WriteString("### Affected Areas\n\n")
		for _, area := range result.AffectedAreas {
			fmt.Fprintf(&b, "- %s\n", area)
		}
		b.WriteString("\n")
	}

	// Warnings
	if len(result.Warnings) > 0 {
		b.WriteString("### Warnings\n\n")
		for _, w := range result.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
		b.WriteString("\n")
	}

	// Changed Files
	b.WriteString("## Changed Files\n\n")
	fmt.Fprintf(&b, "| File | Status | +/- | Language |\n")
	fmt.Fprintf(&b, "|------|--------|-----|----------|\n")
	for _, f := range pr.ChangedFiles {
		flags := fileFlags(f)
		name := f.Path
		if flags != "" {
			name = fmt.Sprintf("%s %s", f.Path, flags)
		}
		fmt.Fprintf(&b, "| %s | %s | +%d/-%d | %s |\n",
			name, f.Status, f.Additions, f.Deletions, f.Language)
	}
	b.WriteString("\n")

	_, err := io.WriteString(w, b.String())
	return err
}

func fileFlags(f domain.ChangedFile) string {
	var flags []string
	if f.IsTest {
		flags = append(flags, "test")
	}
	if f.IsConfig {
		flags = append(flags, "config")
	}
	if f.IsGenerated {
		flags = append(flags, "generated")
	}
	if len(flags) == 0 {
		return ""
	}
	return "(" + strings.Join(flags, ", ") + ")"
}
