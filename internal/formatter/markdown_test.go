package formatter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/formatter"
)

func TestFormatMarkdownContainsTitle(t *testing.T) {
	var buf bytes.Buffer
	err := formatter.FormatMarkdown(&buf, "github", testPR(), testResult())
	if err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Add OAuth2 login") {
		t.Error("output missing PR title header")
	}
}

func TestFormatMarkdownContainsMetadata(t *testing.T) {
	var buf bytes.Buffer
	err := formatter.FormatMarkdown(&buf, "github", testPR(), testResult())
	if err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	checks := []string{
		"owner/repo",
		"#42",
		"dev",
		"feature/oauth",
		"main",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestFormatMarkdownContainsAnalysis(t *testing.T) {
	var buf bytes.Buffer
	err := formatter.FormatMarkdown(&buf, "github", testPR(), testResult())
	if err != nil {
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
	err := formatter.FormatMarkdown(&buf, "github", testPR(), testResult())
	if err != nil {
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
	err := formatter.FormatMarkdown(&buf, "github", testPR(), testResult())
	if err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "New authentication flow") {
		t.Error("output missing warning")
	}
}

func TestFormatMarkdownNoDescription(t *testing.T) {
	pr := domain.PullRequest{
		Repository: "o/r",
		ID:         "1",
		Title:      "Quick fix",
	}
	result := domain.AnalysisResult{
		ChangeType: domain.ChangeTypeBugfix,
		RiskLevel:  domain.RiskLevelLow,
	}
	var buf bytes.Buffer
	err := formatter.FormatMarkdown(&buf, "github", pr, result)
	if err != nil {
		t.Fatalf("FormatMarkdown: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "### Description") {
		t.Error("should not include Description section when description is empty")
	}
}
