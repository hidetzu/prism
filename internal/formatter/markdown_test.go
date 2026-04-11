package formatter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/formatter"
	"github.com/hidetzu/prism/pkg/prism"
)

func TestFormatMarkdownContainsTitle(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, testResult()); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	if !strings.Contains(buf.String(), "# Add OAuth2 login") {
		t.Error("output missing PR title header")
	}
}

func TestFormatMarkdownContainsMetadata(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, testResult()); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	checks := []string{
		"owner/repo",
		"#42",
		"dev",
		"feature/oauth",
		"main",
		"https://github.com/owner/repo/pull/42",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestFormatMarkdownContainsAnalysis(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, testResult()); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "feature") {
		t.Error("output missing change type")
	}
	if !strings.Contains(out, "medium") {
		t.Error("output missing risk level")
	}
	if !strings.Contains(out, "security") {
		t.Error("output missing review axis")
	}
}

func TestFormatMarkdownChangedFiles(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, testResult()); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "internal/auth/oauth.go") {
		t.Error("output missing changed file path")
	}
	if !strings.Contains(out, "(test)") {
		t.Error("output missing test flag on test file")
	}
}

func TestFormatMarkdownWarnings(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, testResult()); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	if !strings.Contains(buf.String(), "New authentication flow") {
		t.Error("output missing warning")
	}
}

func TestFormatMarkdownMinimalResult(t *testing.T) {
	result := prism.Result{
		PR: prism.PRInfo{
			Provider:   "github",
			Repository: "o/r",
			ID:         "1",
			Title:      "Quick fix",
		},
		Analysis: prism.AnalysisResult{
			ChangeType: "bugfix",
			RiskLevel:  "low",
		},
	}
	var buf bytes.Buffer
	if err := formatter.FormatMarkdown(&buf, result); err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Quick fix") {
		t.Error("output missing title")
	}
	if !strings.Contains(out, "bugfix") {
		t.Error("output missing change type")
	}
}
