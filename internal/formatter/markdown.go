package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/hidetzu/prism/pkg/prism"
)

// FormatMarkdown writes a prism.Result as human-readable Markdown to w.
func FormatMarkdown(w io.Writer, result prism.Result) error {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "# %s\n\n", result.PR.Title)

	// Metadata
	b.WriteString("## Pull Request\n\n")
	fmt.Fprintf(&b, "| Field | Value |\n")
	fmt.Fprintf(&b, "|-------|-------|\n")
	fmt.Fprintf(&b, "| Repository | %s |\n", result.PR.Repository)
	fmt.Fprintf(&b, "| PR | #%s |\n", result.PR.ID)
	fmt.Fprintf(&b, "| Author | %s |\n", result.PR.Author)
	fmt.Fprintf(&b, "| Branch | %s -> %s |\n", result.PR.SourceBranch, result.PR.TargetBranch)
	fmt.Fprintf(&b, "| Provider | %s |\n", result.PR.Provider)
	if result.PR.URL != "" {
		fmt.Fprintf(&b, "| URL | %s |\n", result.PR.URL)
	}
	b.WriteString("\n")

	// Analysis
	b.WriteString("## Analysis\n\n")
	fmt.Fprintf(&b, "- **Change Type:** %s\n", result.Analysis.ChangeType)
	fmt.Fprintf(&b, "- **Risk Level:** %s\n", result.Analysis.RiskLevel)
	if result.Analysis.Summary != "" {
		fmt.Fprintf(&b, "- **Summary:** %s\n", result.Analysis.Summary)
	}
	b.WriteString("\n")

	// Review Axes
	if len(result.Analysis.ReviewAxes) > 0 {
		b.WriteString("### Review Axes\n\n")
		for _, axis := range result.Analysis.ReviewAxes {
			fmt.Fprintf(&b, "- %s\n", axis)
		}
		b.WriteString("\n")
	}

	// Affected Areas
	if len(result.Analysis.AffectedAreas) > 0 {
		b.WriteString("### Affected Areas\n\n")
		for _, area := range result.Analysis.AffectedAreas {
			fmt.Fprintf(&b, "- %s\n", area)
		}
		b.WriteString("\n")
	}

	// Warnings
	if len(result.Analysis.Warnings) > 0 {
		b.WriteString("### Warnings\n\n")
		for _, warn := range result.Analysis.Warnings {
			fmt.Fprintf(&b, "- %s\n", warn)
		}
		b.WriteString("\n")
	}

	// Changed Files
	b.WriteString("## Changed Files\n\n")
	fmt.Fprintf(&b, "| File | Status | +/- | Language |\n")
	fmt.Fprintf(&b, "|------|--------|-----|----------|\n")
	for _, f := range result.Files {
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

func fileFlags(f prism.ChangedFile) string {
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
