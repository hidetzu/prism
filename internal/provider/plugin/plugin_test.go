package plugin_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider/plugin"
)

// TestHelperProcess is re-executed by tests as a fake plugin binary.
// It is not a real test — it exits immediately when not invoked by the test harness.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("PRISM_TEST_PLUGIN") == "" {
		return
	}
	behavior := os.Getenv("PRISM_TEST_PLUGIN")
	switch behavior {
	case "valid":
		fmt.Print(`{
			"version": "1",
			"provider": "dummy",
			"repository": "owner/repo",
			"id": "99",
			"title": "Test PR",
			"author": "tester",
			"source_branch": "feature",
			"target_branch": "main",
			"description": "A test pull request",
			"changed_files": [
				{
					"path": "internal/handler.go",
					"status": "modified",
					"additions": 10,
					"deletions": 3,
					"patch": "@@ -1,3 +1,10 @@"
				},
				{
					"path": "internal/handler_test.go",
					"status": "added",
					"additions": 30,
					"deletions": 0,
					"patch": "@@ -0,0 +1,30 @@"
				}
			]
		}`)
	case "invalid_json":
		fmt.Print(`{not valid json`)
	case "wrong_version":
		fmt.Print(`{"version": "99", "provider": "dummy"}`)
	case "stderr_only":
		fmt.Fprint(os.Stderr, "debug info")
		os.Exit(1)
	}
	os.Exit(0)
}

func helperBinary(t *testing.T) string {
	t.Helper()
	return os.Args[0]
}

func newTestProvider(t *testing.T, behavior string) *plugin.Provider {
	t.Helper()
	t.Setenv("PRISM_TEST_PLUGIN", behavior)
	return plugin.NewProvider("dummy", helperBinary(t), "https://example.com/pr/1")
}

func TestFetchPullRequestValid(t *testing.T) {
	p := newTestProvider(t, "valid")
	pr, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "dummy"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.Repository != "owner/repo" {
		t.Errorf("Repository = %q, want %q", pr.Repository, "owner/repo")
	}
	if pr.ID != "99" {
		t.Errorf("ID = %q, want %q", pr.ID, "99")
	}
	if pr.Title != "Test PR" {
		t.Errorf("Title = %q, want %q", pr.Title, "Test PR")
	}
	if pr.Author != "tester" {
		t.Errorf("Author = %q, want %q", pr.Author, "tester")
	}
	if pr.SourceBranch != "feature" {
		t.Errorf("SourceBranch = %q, want %q", pr.SourceBranch, "feature")
	}
	if pr.TargetBranch != "main" {
		t.Errorf("TargetBranch = %q, want %q", pr.TargetBranch, "main")
	}
	if len(pr.ChangedFiles) != 2 {
		t.Fatalf("ChangedFiles length = %d, want 2", len(pr.ChangedFiles))
	}

	f := pr.ChangedFiles[0]
	if f.Path != "internal/handler.go" {
		t.Errorf("file[0].Path = %q", f.Path)
	}
	if f.Language != "Go" {
		t.Errorf("file[0].Language = %q, want %q", f.Language, "Go")
	}
	if f.IsTest {
		t.Error("file[0].IsTest = true, want false")
	}

	f = pr.ChangedFiles[1]
	if !f.IsTest {
		t.Error("file[1].IsTest = false, want true")
	}
}

func TestFetchPullRequestInvalidJSON(t *testing.T) {
	p := newTestProvider(t, "invalid_json")
	_, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "dummy"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchPullRequestWrongVersion(t *testing.T) {
	p := newTestProvider(t, "wrong_version")
	_, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "dummy"})
	if err == nil {
		t.Fatal("expected error for wrong protocol version")
	}
}

func TestFetchPullRequestPluginFailure(t *testing.T) {
	p := newTestProvider(t, "stderr_only")
	_, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "dummy"})
	if err == nil {
		t.Fatal("expected error for plugin failure")
	}
}

func TestFetchPullRequestTimeout(t *testing.T) {
	// Use a command that sleeps, then cancel quickly.
	p := plugin.NewProvider("dummy", "sleep", "10")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := p.FetchPullRequest(ctx, domain.PRRef{Provider: "dummy"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFetchPullRequestBinaryNotFound(t *testing.T) {
	p := plugin.NewProvider("dummy", "/nonexistent/binary", "https://example.com/pr/1")
	_, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "dummy"})
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestParseReturnsMinimalPRRef(t *testing.T) {
	p := plugin.NewProvider("codecommit", "/usr/bin/true", "https://example.com/pr/1")
	ref, err := p.Parse("https://example.com/pr/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Provider != "codecommit" {
		t.Errorf("Provider = %q, want %q", ref.Provider, "codecommit")
	}
}

func TestPluginDiscoveryOnPATH(t *testing.T) {
	// Create a dummy plugin script on PATH.
	dir := t.TempDir()
	script := filepath.Join(dir, "prism-provider-testplugin")
	content := `#!/bin/sh
echo '{"version":"1","provider":"testplugin","repository":"r","id":"1","title":"t","author":"a","source_branch":"s","target_branch":"m","description":"d","changed_files":[]}'
`
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := exec.LookPath(script)
	if err != nil {
		// Add dir to PATH for lookup
		t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
		path, err = exec.LookPath("prism-provider-testplugin")
		if err != nil {
			t.Fatalf("LookPath: %v", err)
		}
	}

	p := plugin.NewProvider("testplugin", path, "https://example.com/pr/1")
	pr, err := p.FetchPullRequest(context.Background(), domain.PRRef{Provider: "testplugin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.ID != "1" {
		t.Errorf("ID = %q, want %q", pr.ID, "1")
	}
}
