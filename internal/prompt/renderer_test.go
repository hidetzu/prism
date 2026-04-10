package prompt_test

import (
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/prompt"
)

func samplePR() domain.PullRequest {
	return domain.PullRequest{
		Repository:   "owner/repo",
		ID:           "42",
		Title:        "Add OAuth2 login",
		Author:       "dev",
		SourceBranch: "feature/oauth",
		TargetBranch: "main",
		Description:  "Implements OAuth2 login flow",
		ChangedFiles: []domain.ChangedFile{
			{
				Path:      "internal/auth/oauth.go",
				Status:    domain.FileStatusAdded,
				Additions: 120,
				Language:  "Go",
				Patch:     "@@ -0,0 +1,120 @@\n+package auth",
			},
			{
				Path:      "internal/auth/oauth_test.go",
				Status:    domain.FileStatusAdded,
				Additions: 80,
				Language:  "Go",
				IsTest:    true,
				Patch:     "@@ -0,0 +1,80 @@\n+package auth_test",
			},
		},
	}
}

func sampleResult() domain.AnalysisResult {
	return domain.AnalysisResult{
		ChangeType:    domain.ChangeTypeFeature,
		RiskLevel:     domain.RiskLevelMedium,
		AffectedAreas: []string{"auth"},
		ReviewAxes:    []domain.ReviewAxis{domain.ReviewAxisSecurity, domain.ReviewAxisErrorHandling},
		RelatedFiles:  []string{"internal/auth/middleware.go"},
		Warnings:      []string{"New authentication flow"},
		Summary:       "Adds OAuth2 login flow with tests",
	}
}

func TestRenderLight(t *testing.T) {
	bundle := prompt.Render(domain.PromptModeLight, samplePR(), sampleResult(), "en")

	if bundle.Mode != domain.PromptModeLight {
		t.Errorf("Mode = %q, want %q", bundle.Mode, domain.PromptModeLight)
	}
	if bundle.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}

	user := bundle.UserPrompt
	checks := []string{
		"Add OAuth2 login",
		"owner/repo",
		"#42",
		"feature",
		"medium",
		"security",
		"error handling",
		"New authentication flow",
		"internal/auth/oauth.go",
	}
	for _, c := range checks {
		if !strings.Contains(user, c) {
			t.Errorf("light UserPrompt missing %q", c)
		}
	}

	// Light mode should NOT include patches.
	if strings.Contains(user, "```diff") {
		t.Error("light mode should not include diff patches")
	}
}

func TestRenderDetailed(t *testing.T) {
	bundle := prompt.Render(domain.PromptModeDetailed, samplePR(), sampleResult(), "en")

	if bundle.Mode != domain.PromptModeDetailed {
		t.Errorf("Mode = %q, want %q", bundle.Mode, domain.PromptModeDetailed)
	}
	if bundle.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}

	user := bundle.UserPrompt
	checks := []string{
		"Add OAuth2 login",
		"owner/repo",
		"Implements OAuth2 login flow",
		"feature",
		"medium",
		"auth",
		"security",
		"error handling",
		"New authentication flow",
		"internal/auth/oauth.go",
		"```diff",
		"(test)",
	}
	for _, c := range checks {
		if !strings.Contains(user, c) {
			t.Errorf("detailed UserPrompt missing %q", c)
		}
	}
}

func TestRenderCross(t *testing.T) {
	bundle := prompt.Render(domain.PromptModeCross, samplePR(), sampleResult(), "en")

	if bundle.Mode != domain.PromptModeCross {
		t.Errorf("Mode = %q, want %q", bundle.Mode, domain.PromptModeCross)
	}
	if bundle.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}

	user := bundle.UserPrompt
	checks := []string{
		"Add OAuth2 login",
		"owner/repo",
		"Module Structure",
		"internal/auth",
		"[source]",
		"[test]",
		"Cross-File Relationships",
		"security",
	}
	for _, c := range checks {
		if !strings.Contains(user, c) {
			t.Errorf("cross UserPrompt missing %q", c)
		}
	}

	// Cross mode should NOT include patches.
	if strings.Contains(user, "```diff") {
		t.Error("cross mode should not include diff patches")
	}

	// Should include test ↔ source pairs.
	if !strings.Contains(user, "Test ↔ Source pairs") {
		t.Error("cross mode should show test/source pairs")
	}
}

func TestRenderLightJA(t *testing.T) {
	bundle := prompt.Render(domain.PromptModeLight, samplePR(), sampleResult(), "ja")

	if !strings.Contains(bundle.SystemPrompt, "コードレビュアー") {
		t.Error("JA system prompt should contain Japanese text")
	}
}

func TestRenderDefaultIsLight(t *testing.T) {
	bundle := prompt.Render("", samplePR(), sampleResult(), "en")
	if bundle.Mode != domain.PromptModeLight {
		t.Errorf("default mode = %q, want %q", bundle.Mode, domain.PromptModeLight)
	}
}
