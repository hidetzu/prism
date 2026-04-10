package usecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/formatter"
	"github.com/hidetzu/prism/internal/usecase"
)

func TestAnalyzeJSON(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{
		Format: "json",
	}, &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	var out formatter.Output
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Provider != "github" {
		t.Errorf("provider = %q, want %q", out.Provider, "github")
	}
	if out.PullRequest.Title != "Fix null pointer in handler" {
		t.Errorf("title = %q", out.PullRequest.Title)
	}
	if out.Analysis.ChangeType != "bugfix" {
		t.Errorf("change_type = %q, want %q", out.Analysis.ChangeType, "bugfix")
	}
}

func TestAnalyzeMarkdown(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{
		Format: "markdown",
	}, &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	out := buf.String()
	if len(out) == 0 {
		t.Error("empty markdown output")
	}
}

func TestAnalyzeText(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{
		Format: "text",
	}, &buf)
	if err != nil {
		t.Fatalf("Analyze text: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Fix null pointer in handler") {
		t.Error("text output missing title")
	}
	if !strings.Contains(out, "bugfix") {
		t.Error("text output missing change type")
	}
}

func TestAnalyzeInvalidFormat(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{
		Format: "xml",
	}, &buf)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestAnalyzeProviderError(t *testing.T) {
	p := &mockProvider{err: errMock}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{}, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAnalyzeClassifiesDocsOnly(t *testing.T) {
	pr := domain.PullRequest{
		Repository: "o/r",
		ID:         "1",
		Title:      "Update docs",
		ChangedFiles: []domain.ChangedFile{
			{Path: "README.md", Status: domain.FileStatusModified, Additions: 5},
		},
	}
	p := &mockProvider{pr: pr}
	var buf bytes.Buffer

	err := usecase.Analyze(context.Background(), p, sampleRef(), usecase.AnalyzeOptions{Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	var out formatter.Output
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Analysis.ChangeType != "docs-only" {
		t.Errorf("change_type = %q, want %q", out.Analysis.ChangeType, "docs-only")
	}
	if out.Analysis.RiskLevel != "low" {
		t.Errorf("risk_level = %q, want %q", out.Analysis.RiskLevel, "low")
	}
}
