package formatter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hidetzu/prism/internal/formatter"
	"github.com/hidetzu/prism/pkg/prism"
)

func testResult() prism.Result {
	return prism.Result{
		PR: prism.PRInfo{
			Provider:     "github",
			Repository:   "owner/repo",
			ID:           "42",
			Title:        "Add OAuth2 login",
			Author:       "dev",
			SourceBranch: "feature/oauth",
			TargetBranch: "main",
			URL:          "https://github.com/owner/repo/pull/42",
		},
		Analysis: prism.AnalysisResult{
			ChangeType:    "feature",
			RiskLevel:     "medium",
			AffectedAreas: []string{"auth"},
			ReviewAxes:    []string{"security", "error handling"},
			Warnings:      []string{"New authentication flow"},
			Summary:       "Adds OAuth2 login flow with tests",
		},
		Files: []prism.ChangedFile{
			{
				Path:      "internal/auth/oauth.go",
				Status:    "added",
				Additions: 120,
				Language:  "Go",
			},
			{
				Path:      "internal/auth/oauth_test.go",
				Status:    "added",
				Additions: 80,
				Language:  "Go",
				IsTest:    true,
			},
		},
	}
}

func TestFormatJSONGolden(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatJSON(&buf, testResult()); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	golden := filepath.Join("testdata", "basic.golden.json")
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	got := normalizeJSON(t, buf.Bytes())
	expected := normalizeJSON(t, want)

	if got != expected {
		t.Errorf("JSON output mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

func TestFormatJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatJSON(&buf, testResult()); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var out prism.Result
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.PR.Provider != "github" {
		t.Errorf("pull_request.provider = %q, want %q", out.PR.Provider, "github")
	}
	if out.PR.ID != "42" {
		t.Errorf("pull_request.id = %q, want %q", out.PR.ID, "42")
	}
	if out.PR.URL != "https://github.com/owner/repo/pull/42" {
		t.Errorf("pull_request.url = %q", out.PR.URL)
	}
	if len(out.Files) != 2 {
		t.Errorf("changed_files length = %d, want 2", len(out.Files))
	}
	if out.Analysis.ChangeType != "feature" {
		t.Errorf("analysis.change_type = %q, want %q", out.Analysis.ChangeType, "feature")
	}
	if out.Analysis.RiskLevel != "medium" {
		t.Errorf("analysis.risk_level = %q, want %q", out.Analysis.RiskLevel, "medium")
	}
}

func TestFormatJSONOmitsDescription(t *testing.T) {
	var buf bytes.Buffer
	if err := formatter.FormatJSON(&buf, testResult()); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	if bytes.Contains(buf.Bytes(), []byte("description")) {
		t.Errorf("output should not contain 'description' field, got: %s", buf.String())
	}
}

func normalizeJSON(t *testing.T, data []byte) string {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("normalize JSON: %v", err)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return string(out)
}
