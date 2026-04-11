package prism

import (
	"testing"

	"github.com/hidetzu/prism/internal/domain"
)

func TestBuildPRURLGitHub(t *testing.T) {
	url := buildPRURL("github", "owner/repo", "123")
	want := "https://github.com/owner/repo/pull/123"
	if url != want {
		t.Errorf("buildPRURL = %q, want %q", url, want)
	}
}

func TestBuildPRURLGitHubEmptyRepo(t *testing.T) {
	url := buildPRURL("github", "", "123")
	if url != "" {
		t.Errorf("buildPRURL with empty repo = %q, want empty", url)
	}
}

func TestBuildPRURLGitHubEmptyID(t *testing.T) {
	url := buildPRURL("github", "owner/repo", "")
	if url != "" {
		t.Errorf("buildPRURL with empty id = %q, want empty", url)
	}
}

func TestBuildPRURLUnknownProvider(t *testing.T) {
	url := buildPRURL("codecommit", "my-repo", "42")
	if url != "" {
		t.Errorf("buildPRURL with unknown provider = %q, want empty", url)
	}
}

func TestBuildPRURLEmptyProvider(t *testing.T) {
	url := buildPRURL("", "owner/repo", "123")
	if url != "" {
		t.Errorf("buildPRURL with empty provider = %q, want empty", url)
	}
}

func TestBuildResultBasic(t *testing.T) {
	pr := domain.PullRequest{
		Repository:   "owner/repo",
		ID:           "42",
		Title:        "Fix bug",
		Author:       "dev",
		SourceBranch: "fix/bug",
		TargetBranch: "main",
		Description:  "fixes a bug",
		ChangedFiles: []domain.ChangedFile{
			{
				Path:        "handler.go",
				Status:      domain.FileStatusModified,
				Additions:   10,
				Deletions:   5,
				Language:    "Go",
				IsTest:      false,
				IsConfig:    false,
				IsGenerated: false,
				Patch:       "@@ -1,3 +1,10 @@",
			},
		},
	}
	analysis := domain.AnalysisResult{
		ChangeType:    domain.ChangeTypeBugfix,
		RiskLevel:     domain.RiskLevelMedium,
		AffectedAreas: []string{"handler"},
		ReviewAxes: []domain.ReviewAxis{
			domain.ReviewAxisErrorHandling,
			domain.ReviewAxisEdgeCases,
		},
		RelatedFiles: []string{"handler_test.go"},
		Warnings:     []string{"No tests added"},
		Summary:      "bugfix: Fix bug",
	}

	result := buildResult("github", pr, analysis, false)

	// PR info
	if result.PR.Provider != "github" {
		t.Errorf("PR.Provider = %q", result.PR.Provider)
	}
	if result.PR.Repository != "owner/repo" {
		t.Errorf("PR.Repository = %q", result.PR.Repository)
	}
	if result.PR.ID != "42" {
		t.Errorf("PR.ID = %q", result.PR.ID)
	}
	if result.PR.Title != "Fix bug" {
		t.Errorf("PR.Title = %q", result.PR.Title)
	}
	if result.PR.Author != "dev" {
		t.Errorf("PR.Author = %q", result.PR.Author)
	}
	if result.PR.SourceBranch != "fix/bug" {
		t.Errorf("PR.SourceBranch = %q", result.PR.SourceBranch)
	}
	if result.PR.TargetBranch != "main" {
		t.Errorf("PR.TargetBranch = %q", result.PR.TargetBranch)
	}
	if result.PR.URL != "https://github.com/owner/repo/pull/42" {
		t.Errorf("PR.URL = %q", result.PR.URL)
	}

	// Analysis
	if result.Analysis.ChangeType != "bugfix" {
		t.Errorf("Analysis.ChangeType = %q", result.Analysis.ChangeType)
	}
	if result.Analysis.RiskLevel != "medium" {
		t.Errorf("Analysis.RiskLevel = %q", result.Analysis.RiskLevel)
	}
	if len(result.Analysis.AffectedAreas) != 1 || result.Analysis.AffectedAreas[0] != "handler" {
		t.Errorf("Analysis.AffectedAreas = %v", result.Analysis.AffectedAreas)
	}
	if len(result.Analysis.ReviewAxes) != 2 {
		t.Fatalf("Analysis.ReviewAxes len = %d, want 2", len(result.Analysis.ReviewAxes))
	}
	if result.Analysis.ReviewAxes[0] != "error handling" {
		t.Errorf("Analysis.ReviewAxes[0] = %q", result.Analysis.ReviewAxes[0])
	}
	if result.Analysis.ReviewAxes[1] != "edge cases" {
		t.Errorf("Analysis.ReviewAxes[1] = %q", result.Analysis.ReviewAxes[1])
	}
	if result.Analysis.Summary != "bugfix: Fix bug" {
		t.Errorf("Analysis.Summary = %q", result.Analysis.Summary)
	}

	// Files
	if len(result.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(result.Files))
	}
	f := result.Files[0]
	if f.Path != "handler.go" {
		t.Errorf("Files[0].Path = %q", f.Path)
	}
	if f.Status != "modified" {
		t.Errorf("Files[0].Status = %q", f.Status)
	}
	if f.Additions != 10 {
		t.Errorf("Files[0].Additions = %d", f.Additions)
	}
	if f.Deletions != 5 {
		t.Errorf("Files[0].Deletions = %d", f.Deletions)
	}
	if f.Language != "Go" {
		t.Errorf("Files[0].Language = %q", f.Language)
	}
}

func TestBuildResultExcludesPatchByDefault(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "a.go", Status: domain.FileStatusModified, Patch: "@@ diff content"},
		},
	}
	result := buildResult("github", pr, domain.AnalysisResult{}, false)

	if len(result.Files) != 1 {
		t.Fatal("expected 1 file")
	}
	if result.Files[0].Patch != "" {
		t.Errorf("Patch = %q, want empty when IncludePatches=false", result.Files[0].Patch)
	}
}

func TestBuildResultIncludesPatchWhenRequested(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "a.go", Status: domain.FileStatusModified, Patch: "@@ diff content"},
		},
	}
	result := buildResult("github", pr, domain.AnalysisResult{}, true)

	if len(result.Files) != 1 {
		t.Fatal("expected 1 file")
	}
	if result.Files[0].Patch != "@@ diff content" {
		t.Errorf("Patch = %q, want %q", result.Files[0].Patch, "@@ diff content")
	}
}

func TestBuildResultEmptyChangedFiles(t *testing.T) {
	pr := domain.PullRequest{
		Repository: "owner/repo",
		ID:         "1",
	}
	result := buildResult("github", pr, domain.AnalysisResult{}, false)

	if result.Files == nil {
		t.Error("Files should not be nil (should be empty slice)")
	}
	if len(result.Files) != 0 {
		t.Errorf("Files len = %d, want 0", len(result.Files))
	}
}

func TestBuildResultFileFlags(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "x_test.go", Status: domain.FileStatusAdded, IsTest: true},
			{Path: "config.yaml", Status: domain.FileStatusModified, IsConfig: true},
			{Path: "api.gen.go", Status: domain.FileStatusModified, IsGenerated: true},
		},
	}
	result := buildResult("github", pr, domain.AnalysisResult{}, false)

	if !result.Files[0].IsTest {
		t.Error("Files[0].IsTest should be true")
	}
	if !result.Files[1].IsConfig {
		t.Error("Files[1].IsConfig should be true")
	}
	if !result.Files[2].IsGenerated {
		t.Error("Files[2].IsGenerated should be true")
	}
	if result.Files[0].Status != "added" {
		t.Errorf("Files[0].Status = %q, want added", result.Files[0].Status)
	}
}

func TestBuildResultEmptyReviewAxes(t *testing.T) {
	pr := domain.PullRequest{}
	result := buildResult("github", pr, domain.AnalysisResult{}, false)

	if result.Analysis.ReviewAxes == nil {
		t.Error("ReviewAxes should not be nil (should be empty slice)")
	}
	if len(result.Analysis.ReviewAxes) != 0 {
		t.Errorf("ReviewAxes len = %d, want 0", len(result.Analysis.ReviewAxes))
	}
}
