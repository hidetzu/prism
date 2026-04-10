package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider/github"
)

func setupTestServer(t *testing.T) (*httptest.Server, domain.PRRef) {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /repos/owner/repo/pulls/42", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Error("missing Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"number": 42,
			"title": "Add feature X",
			"body": "This adds feature X",
			"user": {"login": "dev"},
			"head": {"ref": "feature-x"},
			"base": {
				"ref": "main",
				"repo": {"full_name": "owner/repo"}
			}
		}`)
	})

	mux.HandleFunc("GET /repos/owner/repo/pulls/42/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[
			{
				"filename": "internal/handler.go",
				"status": "modified",
				"additions": 20,
				"deletions": 5,
				"patch": "@@ -1,5 +1,20 @@"
			},
			{
				"filename": "internal/handler_test.go",
				"status": "added",
				"additions": 50,
				"deletions": 0,
				"patch": "@@ -0,0 +1,50 @@"
			},
			{
				"filename": "config.yaml",
				"status": "modified",
				"additions": 1,
				"deletions": 0,
				"patch": "@@ -1 +1,2 @@"
			},
			{
				"filename": "api.gen.go",
				"status": "modified",
				"additions": 100,
				"deletions": 80,
				"patch": ""
			}
		]`)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	ref := domain.PRRef{
		Provider: "github",
		Owner:    "owner",
		Repo:     "repo",
		Number:   42,
	}

	return server, ref
}

func TestFetchPullRequest(t *testing.T) {
	server, ref := setupTestServer(t)

	p := github.NewProviderWithClient(server.Client(), "test-token", server.URL)
	pr, err := p.FetchPullRequest(context.Background(), ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr.Repository != "owner/repo" {
		t.Errorf("Repository = %q, want %q", pr.Repository, "owner/repo")
	}
	if pr.ID != "42" {
		t.Errorf("ID = %q, want %q", pr.ID, "42")
	}
	if pr.Title != "Add feature X" {
		t.Errorf("Title = %q, want %q", pr.Title, "Add feature X")
	}
	if pr.Author != "dev" {
		t.Errorf("Author = %q, want %q", pr.Author, "dev")
	}
	if pr.SourceBranch != "feature-x" {
		t.Errorf("SourceBranch = %q, want %q", pr.SourceBranch, "feature-x")
	}
	if pr.TargetBranch != "main" {
		t.Errorf("TargetBranch = %q, want %q", pr.TargetBranch, "main")
	}
	if pr.Description != "This adds feature X" {
		t.Errorf("Description = %q, want %q", pr.Description, "This adds feature X")
	}

	if len(pr.ChangedFiles) != 4 {
		t.Fatalf("ChangedFiles length = %d, want 4", len(pr.ChangedFiles))
	}

	// handler.go
	f := pr.ChangedFiles[0]
	if f.Path != "internal/handler.go" {
		t.Errorf("file[0].Path = %q", f.Path)
	}
	if f.Status != domain.FileStatusModified {
		t.Errorf("file[0].Status = %q", f.Status)
	}
	if f.Language != "Go" {
		t.Errorf("file[0].Language = %q, want %q", f.Language, "Go")
	}
	if f.IsTest {
		t.Error("file[0].IsTest = true, want false")
	}

	// handler_test.go
	f = pr.ChangedFiles[1]
	if f.Status != domain.FileStatusAdded {
		t.Errorf("file[1].Status = %q", f.Status)
	}
	if !f.IsTest {
		t.Error("file[1].IsTest = false, want true")
	}

	// config.yaml
	f = pr.ChangedFiles[2]
	if !f.IsConfig {
		t.Error("file[2].IsConfig = false, want true")
	}
	if f.Language != "YAML" {
		t.Errorf("file[2].Language = %q, want %q", f.Language, "YAML")
	}

	// api.gen.go
	f = pr.ChangedFiles[3]
	if !f.IsGenerated {
		t.Error("file[3].IsGenerated = false, want true")
	}
}

func TestFetchPullRequestAuthHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/repos/o/r/pulls/1" {
			_, _ = fmt.Fprint(w, `{"number":1,"title":"t","body":"","user":{"login":"u"},"head":{"ref":"b"},"base":{"ref":"main","repo":{"full_name":"o/r"}}}`)
		} else {
			_, _ = fmt.Fprint(w, `[]`)
		}
	}))
	t.Cleanup(server.Close)

	ref := domain.PRRef{Provider: "github", Owner: "o", Repo: "r", Number: 1}
	p := github.NewProviderWithClient(server.Client(), "my-token", server.URL)
	_, err := p.FetchPullRequest(context.Background(), ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer my-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer my-token")
	}
}

func TestFetchPullRequestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"Not Found"}`)
	}))
	t.Cleanup(server.Close)

	ref := domain.PRRef{Provider: "github", Owner: "o", Repo: "r", Number: 999}
	p := github.NewProviderWithClient(server.Client(), "", server.URL)
	_, err := p.FetchPullRequest(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestFetchPullRequestContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ref := domain.PRRef{Provider: "github", Owner: "o", Repo: "r", Number: 1}
	p := github.NewProviderWithClient(server.Client(), "", server.URL)
	_, err := p.FetchPullRequest(ctx, ref)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFetchFilesPaginationLimit(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/repos/o/r/pulls/1" {
			_, _ = fmt.Fprint(w, `{"number":1,"title":"t","body":"","user":{"login":"u"},"head":{"ref":"b"},"base":{"ref":"main","repo":{"full_name":"o/r"}}}`)
			return
		}
		pageCount++
		_, _ = fmt.Fprint(w, "[")
		for i := 0; i < 100; i++ {
			if i > 0 {
				_, _ = fmt.Fprint(w, ",")
			}
			_, _ = fmt.Fprintf(w, `{"filename":"file%d_%d.go","status":"added","additions":1,"deletions":0,"patch":""}`, pageCount, i)
		}
		_, _ = fmt.Fprint(w, "]")
	}))
	t.Cleanup(server.Close)

	ref := domain.PRRef{Provider: "github", Owner: "o", Repo: "r", Number: 1}
	p := github.NewProviderWithClient(server.Client(), "", server.URL)
	_, err := p.FetchPullRequest(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error for too many pages")
	}
}
