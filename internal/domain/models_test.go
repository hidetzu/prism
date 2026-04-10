package domain_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/domain"
)

func TestPRRef(t *testing.T) {
	ref := domain.PRRef{
		Provider: "github",
		Owner:    "owner",
		Repo:     "repo",
		Number:   42,
	}

	if ref.Provider != "github" {
		t.Errorf("Provider = %q, want %q", ref.Provider, "github")
	}
	if ref.Owner != "owner" {
		t.Errorf("Owner = %q, want %q", ref.Owner, "owner")
	}
	if ref.Repo != "repo" {
		t.Errorf("Repo = %q, want %q", ref.Repo, "repo")
	}
	if ref.Number != 42 {
		t.Errorf("Number = %d, want %d", ref.Number, 42)
	}
}

func TestPullRequest(t *testing.T) {
	pr := domain.PullRequest{
		Repository:   "owner/repo",
		ID:           "123",
		Title:        "Add feature X",
		Author:       "dev",
		SourceBranch: "feature-x",
		TargetBranch: "main",
		Description:  "Adds feature X",
		ChangedFiles: []domain.ChangedFile{
			{
				Path:      "src/feature.go",
				Status:    domain.FileStatusAdded,
				Additions: 50,
				Deletions: 0,
				Language:  "Go",
			},
		},
	}

	if pr.Repository != "owner/repo" {
		t.Errorf("Repository = %q, want %q", pr.Repository, "owner/repo")
	}
	if len(pr.ChangedFiles) != 1 {
		t.Fatalf("ChangedFiles length = %d, want 1", len(pr.ChangedFiles))
	}
	if pr.ChangedFiles[0].Status != domain.FileStatusAdded {
		t.Errorf("Status = %q, want %q", pr.ChangedFiles[0].Status, domain.FileStatusAdded)
	}
}

func TestChangedFileFlags(t *testing.T) {
	f := domain.ChangedFile{
		Path:        "internal/handler_test.go",
		Status:      domain.FileStatusModified,
		Additions:   10,
		Deletions:   5,
		Language:    "Go",
		IsTest:      true,
		IsConfig:    false,
		IsGenerated: false,
	}

	if !f.IsTest {
		t.Error("IsTest = false, want true")
	}
	if f.IsConfig {
		t.Error("IsConfig = true, want false")
	}
}

func TestAnalysisResult(t *testing.T) {
	result := domain.AnalysisResult{
		ChangeType:    domain.ChangeTypeFeature,
		RiskLevel:     domain.RiskLevelMedium,
		AffectedAreas: []string{"auth"},
		ReviewAxes:    []domain.ReviewAxis{domain.ReviewAxisSecurity, domain.ReviewAxisErrorHandling},
		Warnings:      []string{"No tests added"},
		Summary:       "Adds OAuth2 login flow",
	}

	if result.ChangeType != domain.ChangeTypeFeature {
		t.Errorf("ChangeType = %q, want %q", result.ChangeType, domain.ChangeTypeFeature)
	}
	if result.RiskLevel != domain.RiskLevelMedium {
		t.Errorf("RiskLevel = %q, want %q", result.RiskLevel, domain.RiskLevelMedium)
	}
	if len(result.ReviewAxes) != 2 {
		t.Fatalf("ReviewAxes length = %d, want 2", len(result.ReviewAxes))
	}
}

func TestPromptBundle(t *testing.T) {
	bundle := domain.PromptBundle{
		Mode:         domain.PromptModeLight,
		SystemPrompt: "You are a code reviewer.",
		UserPrompt:   "Review this PR.",
	}

	if bundle.Mode != domain.PromptModeLight {
		t.Errorf("Mode = %q, want %q", bundle.Mode, domain.PromptModeLight)
	}
}

func TestChangeTypeConstants(t *testing.T) {
	types := []domain.ChangeType{
		domain.ChangeTypeFeature,
		domain.ChangeTypeBugfix,
		domain.ChangeTypeRefactor,
		domain.ChangeTypeTestOnly,
		domain.ChangeTypeDocsOnly,
		domain.ChangeTypeConfigChange,
		domain.ChangeTypeDependencyUpdate,
		domain.ChangeTypeInfraChange,
	}

	seen := make(map[domain.ChangeType]bool)
	for _, ct := range types {
		if ct == "" {
			t.Error("ChangeType constant is empty")
		}
		if seen[ct] {
			t.Errorf("duplicate ChangeType: %q", ct)
		}
		seen[ct] = true
	}
}

func TestFileStatusConstants(t *testing.T) {
	statuses := []domain.FileStatus{
		domain.FileStatusAdded,
		domain.FileStatusModified,
		domain.FileStatusRemoved,
		domain.FileStatusRenamed,
	}

	seen := make(map[domain.FileStatus]bool)
	for _, s := range statuses {
		if s == "" {
			t.Error("FileStatus constant is empty")
		}
		if seen[s] {
			t.Errorf("duplicate FileStatus: %q", s)
		}
		seen[s] = true
	}
}
