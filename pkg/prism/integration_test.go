package prism

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hidetzu/prism/internal/provider"
)

// setupGitHubMock spins up an httptest server that mimics the GitHub REST API
// for a single PR. It returns the server URL the registry should target.
func setupGitHubMock(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"number": 1,
			"title": "Add retry handling",
			"body": "Adds retry logic for transient failures",
			"user": {"login": "dev"},
			"head": {"ref": "feature/retry"},
			"base": {"ref": "main", "repo": {"full_name": "owner/repo"}}
		}`)
	})

	mux.HandleFunc("GET /repos/owner/repo/pulls/1/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[
			{
				"filename": "internal/payment/client.go",
				"status": "modified",
				"additions": 45,
				"deletions": 3,
				"patch": "@@ -1,3 +1,45 @@"
			},
			{
				"filename": "internal/payment/client_test.go",
				"status": "added",
				"additions": 80,
				"deletions": 0,
				"patch": "@@ -0,0 +1,80 @@"
			}
		]`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// withMockGitHub overrides newRegistry so that the github provider talks to
// the given mock server. The override is reverted on test cleanup.
func withMockGitHub(t *testing.T, server *httptest.Server) {
	t.Helper()
	previous := newRegistry
	newRegistry = func(githubToken string) *provider.Registry {
		return provider.NewRegistryWithGitHubBaseURL(githubToken, server.URL)
	}
	t.Cleanup(func() { newRegistry = previous })
}

func TestAnalyzeHappyPath(t *testing.T) {
	server := setupGitHubMock(t)
	withMockGitHub(t, server)

	result, err := Analyze(context.Background(), AnalyzeOptions{
		PRURL:       "https://github.com/owner/repo/pull/1",
		GitHubToken: "test-token",
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	// PR metadata
	if result.PR.Provider != "github" {
		t.Errorf("PR.Provider = %q, want github", result.PR.Provider)
	}
	if result.PR.Repository != "owner/repo" {
		t.Errorf("PR.Repository = %q, want owner/repo", result.PR.Repository)
	}
	if result.PR.ID != "1" {
		t.Errorf("PR.ID = %q, want 1", result.PR.ID)
	}
	if result.PR.Title != "Add retry handling" {
		t.Errorf("PR.Title = %q", result.PR.Title)
	}
	if result.PR.URL != "https://github.com/owner/repo/pull/1" {
		t.Errorf("PR.URL = %q", result.PR.URL)
	}

	// Analysis was actually run
	if result.Analysis.ChangeType == "" {
		t.Error("Analysis.ChangeType is empty")
	}
	if result.Analysis.RiskLevel == "" {
		t.Error("Analysis.RiskLevel is empty")
	}

	// Files come through with enrichment from internal/fileutil
	if len(result.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(result.Files))
	}
	if result.Files[0].Language != "Go" {
		t.Errorf("Files[0].Language = %q, want Go", result.Files[0].Language)
	}
	if !result.Files[1].IsTest {
		t.Error("Files[1].IsTest = false, want true (test file detection)")
	}

	// IncludePatches=false (default) → patches must be empty
	for i, f := range result.Files {
		if f.Patch != "" {
			t.Errorf("Files[%d].Patch = %q, want empty (IncludePatches=false)", i, f.Patch)
		}
	}
}

func TestAnalyzeHappyPathIncludePatches(t *testing.T) {
	server := setupGitHubMock(t)
	withMockGitHub(t, server)

	result, err := Analyze(context.Background(), AnalyzeOptions{
		PRURL:          "https://github.com/owner/repo/pull/1",
		GitHubToken:    "test-token",
		IncludePatches: true,
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(result.Files))
	}
	for i, f := range result.Files {
		if f.Patch == "" {
			t.Errorf("Files[%d].Patch is empty, want non-empty (IncludePatches=true)", i)
		}
	}
}

func TestAnalyzeExplicitProvider(t *testing.T) {
	server := setupGitHubMock(t)
	withMockGitHub(t, server)

	result, err := Analyze(context.Background(), AnalyzeOptions{
		PRURL:       "https://github.com/owner/repo/pull/1",
		Provider:    "github",
		GitHubToken: "test-token",
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if result.PR.Provider != "github" {
		t.Errorf("PR.Provider = %q, want github", result.PR.Provider)
	}
}

func TestAnalyzeUpstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	withMockGitHub(t, server)

	_, err := Analyze(context.Background(), AnalyzeOptions{
		PRURL:       "https://github.com/owner/repo/pull/1",
		GitHubToken: "test-token",
	})
	if err == nil {
		t.Fatal("expected error for upstream failure")
	}
	if !errors.Is(err, ErrUpstreamFailure) {
		t.Errorf("expected ErrUpstreamFailure, got %v", err)
	}
}

func TestPromptHappyPath(t *testing.T) {
	server := setupGitHubMock(t)
	withMockGitHub(t, server)

	out, err := Prompt(context.Background(), AnalyzeOptions{
		PRURL:       "https://github.com/owner/repo/pull/1",
		GitHubToken: "test-token",
		Mode:        "light",
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if out == "" {
		t.Fatal("Prompt returned empty string")
	}
	// Light mode should include the PR title in the user prompt.
	// We don't assert exact format, just sanity-check that something useful came out.
	if len(out) < 50 {
		t.Errorf("Prompt output too short: %q", out)
	}
}
