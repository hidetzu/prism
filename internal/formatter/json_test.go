package formatter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/formatter"
)

func testPR() domain.PullRequest {
	return domain.PullRequest{
		Repository:   "owner/repo",
		ID:           "42",
		Title:        "Add OAuth2 login",
		Author:       "dev",
		SourceBranch: "feature/oauth",
		TargetBranch: "main",
		Description:  "Adds OAuth2 login flow",
		ChangedFiles: []domain.ChangedFile{
			{
				Path:      "internal/auth/oauth.go",
				Status:    domain.FileStatusAdded,
				Additions: 120,
				Language:  "Go",
			},
			{
				Path:      "internal/auth/oauth_test.go",
				Status:    domain.FileStatusAdded,
				Additions: 80,
				Language:  "Go",
				IsTest:    true,
			},
		},
	}
}

func testResult() domain.AnalysisResult {
	return domain.AnalysisResult{
		ChangeType:    domain.ChangeTypeFeature,
		RiskLevel:     domain.RiskLevelMedium,
		AffectedAreas: []string{"auth"},
		ReviewAxes:    []domain.ReviewAxis{domain.ReviewAxisSecurity, domain.ReviewAxisErrorHandling},
		Warnings:      []string{"New authentication flow"},
		Summary:       "Adds OAuth2 login flow with tests",
	}
}

func TestFormatJSONGolden(t *testing.T) {
	var buf bytes.Buffer
	err := formatter.FormatJSON(&buf, "github", testPR(), testResult())
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	golden := filepath.Join("testdata", "basic.golden.json")
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	// Normalize both by re-encoding to ensure consistent formatting.
	got := normalizeJSON(t, buf.Bytes())
	expected := normalizeJSON(t, want)

	if got != expected {
		t.Errorf("JSON output mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, expected)
	}
}

func TestFormatJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	err := formatter.FormatJSON(&buf, "github", testPR(), testResult())
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var out formatter.Output
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Provider != "github" {
		t.Errorf("provider = %q, want %q", out.Provider, "github")
	}
	if out.PullRequest.ID != "42" {
		t.Errorf("pull_request.id = %q, want %q", out.PullRequest.ID, "42")
	}
	if len(out.ChangedFiles) != 2 {
		t.Errorf("changed_files length = %d, want 2", len(out.ChangedFiles))
	}
	if out.Analysis.ChangeType != "feature" {
		t.Errorf("analysis.change_type = %q, want %q", out.Analysis.ChangeType, "feature")
	}
	if out.Analysis.RiskLevel != "medium" {
		t.Errorf("analysis.risk_level = %q, want %q", out.Analysis.RiskLevel, "medium")
	}
}

func TestFormatJSONNilSlices(t *testing.T) {
	pr := domain.PullRequest{
		Repository: "o/r",
		ID:         "1",
	}
	result := domain.AnalysisResult{
		ChangeType: domain.ChangeTypeBugfix,
		RiskLevel:  domain.RiskLevelLow,
		// All slice fields are nil.
	}

	var buf bytes.Buffer
	err := formatter.FormatJSON(&buf, "github", pr, result)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var out formatter.Output
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Nil slices should become empty arrays, not null.
	if out.Analysis.AffectedAreas == nil {
		t.Error("affected_areas is null, want empty array")
	}
	if out.Analysis.ReviewAxes == nil {
		t.Error("review_axes is null, want empty array")
	}
	if out.Analysis.RelatedFiles == nil {
		t.Error("related_files is null, want empty array")
	}
	if out.Analysis.Warnings == nil {
		t.Error("warnings is null, want empty array")
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
