package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/formatter"
	"github.com/hidetzu/prism/internal/provider"
	ghprovider "github.com/hidetzu/prism/internal/provider/github"
	"github.com/hidetzu/prism/internal/usecase"
)

// setupGitHubMock creates a test HTTP server that mimics GitHub API responses.
func setupGitHubMock(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /repos/owner/repo/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"number": 1,
			"title": "Fix login bug",
			"body": "Fixes #42",
			"user": {"login": "dev"},
			"head": {"ref": "fix/login"},
			"base": {"ref": "main", "repo": {"full_name": "owner/repo"}}
		}`)
	})

	mux.HandleFunc("GET /repos/owner/repo/pulls/1/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[
			{"filename": "internal/auth/login.go", "status": "modified", "additions": 10, "deletions": 3, "patch": "@@ -1,3 +1,10 @@"},
			{"filename": "internal/auth/login_test.go", "status": "modified", "additions": 20, "deletions": 0, "patch": "@@ -1 +1,20 @@"}
		]`)
	})

	return httptest.NewServer(mux)
}

func TestCLIAnalyzeRequiresURL(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"analyze"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "PR URL is required") {
		t.Errorf("error = %q, want 'PR URL is required'", err.Error())
	}
}

func TestCLIAnalyzeInvalidURL(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"analyze", "https://not-github.com/a/b/pull/1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot auto-detect provider") {
		t.Errorf("error = %q, want 'cannot auto-detect provider'", err.Error())
	}
}

func TestCLIAnalyzeWithUnknownProvider(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"analyze", "https://example.com/pr/1", "--provider", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestCLIPromptInvalidMode(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"prompt", "https://github.com/a/b/pull/1", "--mode", "bad"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("error = %q, want 'invalid mode'", err.Error())
	}
}

func TestCLIPromptInvalidLang(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"prompt", "https://github.com/a/b/pull/1", "--lang", "fr"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid language") {
		t.Errorf("error = %q, want 'invalid language'", err.Error())
	}
}

func TestCLIAnalyzeInvalidFormat(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"analyze", "https://github.com/a/b/pull/1", "--format", "xml"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error = %q, want 'invalid format'", err.Error())
	}
}

func TestCLIFetchInvalidFormat(t *testing.T) {
	cmd := rootCmd()
	cmd.SetArgs([]string{"fetch", "https://github.com/a/b/pull/1", "--format", "markdown"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error = %q, want 'invalid format'", err.Error())
	}
}

func TestCLIVersion(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(buf.String(), version) {
		t.Errorf("version output = %q, want to contain %q", buf.String(), version)
	}
}

func TestCLIHelpContainsCommands(t *testing.T) {
	cmd := rootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"analyze", "prompt", "fetch"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing command %q", want)
		}
	}
}

// TestUsecaseIntegrationAnalyzeJSON tests the full pipeline through usecase layer
// with a mock HTTP server, verifying JSON output structure.
func TestUsecaseIntegrationAnalyzeJSON(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	// Use provider with mock server directly.
	ghprov := newTestProvider(t, server)
	ref, err := ghprov.Parse("https://github.com/owner/repo/pull/1")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	var buf bytes.Buffer
	err = usecaseAnalyze(t, ghprov, ref, "json", &buf)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	var out formatter.Output
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Provider != "github" {
		t.Errorf("provider = %q", out.Provider)
	}
	if out.PullRequest.Title != "Fix login bug" {
		t.Errorf("title = %q", out.PullRequest.Title)
	}
	if out.Analysis.ChangeType != "bugfix" {
		t.Errorf("change_type = %q, want bugfix", out.Analysis.ChangeType)
	}
	if len(out.ChangedFiles) != 2 {
		t.Errorf("changed_files = %d, want 2", len(out.ChangedFiles))
	}
}

func TestUsecaseIntegrationAnalyzeMarkdown(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecaseAnalyze(t, ghprov, ref, "markdown", &buf); err != nil {
		t.Fatalf("Analyze markdown: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "# Fix login bug") {
		t.Error("markdown missing title")
	}
	if !strings.Contains(out, "bugfix") {
		t.Error("markdown missing change type")
	}
}

func TestUsecaseIntegrationAnalyzeText(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecaseAnalyze(t, ghprov, ref, "text", &buf); err != nil {
		t.Fatalf("Analyze text: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Fix login bug") {
		t.Error("text missing title")
	}
	if !strings.Contains(out, "bugfix") {
		t.Error("text missing change type")
	}
}

func TestUsecaseIntegrationPromptLight(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecasePrompt(t, ghprov, ref, "light", "text", "en", &buf); err != nil {
		t.Fatalf("Prompt light: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Fix login bug") {
		t.Error("light prompt missing title")
	}
	if strings.Contains(out, "```diff") {
		t.Error("light should not contain diff")
	}
}

func TestUsecaseIntegrationPromptDetailed(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecasePrompt(t, ghprov, ref, "detailed", "text", "en", &buf); err != nil {
		t.Fatalf("Prompt detailed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "```diff") {
		t.Error("detailed should contain diff")
	}
}

func TestUsecaseIntegrationPromptCross(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecasePrompt(t, ghprov, ref, "cross", "text", "en", &buf); err != nil {
		t.Fatalf("Prompt cross: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Module Structure") {
		t.Error("cross missing module structure")
	}
	if !strings.Contains(out, "Cross-File Relationships") {
		t.Error("cross missing relationships")
	}
	if !strings.Contains(out, "Test ↔ Source pairs") {
		t.Error("cross missing test/source pairs")
	}
}

func TestUsecaseIntegrationPromptJA(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecasePrompt(t, ghprov, ref, "light", "text", "ja", &buf); err != nil {
		t.Fatalf("Prompt ja: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "コードレビュアー") {
		t.Error("ja prompt missing Japanese system prompt")
	}
}

func TestUsecaseIntegrationFetchJSON(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecaseFetch(t, ghprov, ref, "json", &buf); err != nil {
		t.Fatalf("Fetch json: %v", err)
	}
	if !strings.Contains(buf.String(), "Fix login bug") {
		t.Error("fetch json missing title")
	}
}

func TestUsecaseIntegrationFetchText(t *testing.T) {
	server := setupGitHubMock(t)
	t.Cleanup(server.Close)

	ghprov := newTestProvider(t, server)
	ref, _ := ghprov.Parse("https://github.com/owner/repo/pull/1")

	var buf bytes.Buffer
	if err := usecaseFetch(t, ghprov, ref, "text", &buf); err != nil {
		t.Fatalf("Fetch text: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Repository: owner/repo") {
		t.Error("fetch text missing repo")
	}
}

// Helpers to call usecases with test provider.

func newTestProvider(t *testing.T, server *httptest.Server) *ghProviderForTest {
	t.Helper()
	return &ghProviderForTest{server: server}
}

type ghProviderForTest struct {
	server *httptest.Server
}

func (p *ghProviderForTest) Parse(input string) (domain.PRRef, error) {
	return ghprovider.Parse(input)
}

func (p *ghProviderForTest) FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	prov := ghprovider.NewProviderWithClient(p.server.Client(), "", p.server.URL)
	return prov.FetchPullRequest(ctx, ref)
}

func usecaseAnalyze(t *testing.T, p provider.Provider, ref domain.PRRef, format string, buf *bytes.Buffer) error {
	t.Helper()
	return usecase.Analyze(context.Background(), p, ref, usecase.AnalyzeOptions{Format: format}, buf)
}

func usecasePrompt(t *testing.T, p provider.Provider, ref domain.PRRef, mode, format, lang string, buf *bytes.Buffer) error {
	t.Helper()
	return usecase.Prompt(context.Background(), p, ref, usecase.PromptOptions{Mode: mode, Format: format, Lang: lang}, buf)
}

func usecaseFetch(t *testing.T, p provider.Provider, ref domain.PRRef, format string, buf *bytes.Buffer) error {
	t.Helper()
	return usecase.Fetch(context.Background(), p, ref, usecase.FetchOptions{Format: format}, buf)
}
