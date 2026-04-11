package formatter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/formatter"
)

func TestFormatTextContainsTitle(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatText(&buf, testResult()); err != nil {
		t.Fatalf("FormatText: %v", err)
	}
	if !strings.Contains(buf.String(), "Add OAuth2 login") {
		t.Error("output missing PR title")
	}
}

func TestFormatTextContainsMetadata(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatText(&buf, testResult()); err != nil {
		t.Fatalf("FormatText: %v", err)
	}
	out := buf.String()
	checks := []string{
		"owner/repo",
		"#42",
		"dev",
		"feature/oauth",
		"main",
		"github",
		"https://github.com/owner/repo/pull/42",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q", c)
		}
	}
}

func TestFormatTextContainsAnalysis(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatText(&buf, testResult()); err != nil {
		t.Fatalf("FormatText: %v", err)
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

func TestFormatTextChangedFiles(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatText(&buf, testResult()); err != nil {
		t.Fatalf("FormatText: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "internal/auth/oauth.go") {
		t.Error("output missing file path")
	}
	if !strings.Contains(out, "(test)") {
		t.Error("output missing test flag")
	}
}

func TestFormatTextWarnings(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatText(&buf, testResult()); err != nil {
		t.Fatalf("FormatText: %v", err)
	}
	if !strings.Contains(buf.String(), "New authentication flow") {
		t.Error("output missing warning")
	}
}
