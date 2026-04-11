package provider

import (
	"testing"
)

func TestDetectProviderGitHub(t *testing.T) {
	name, err := detectProvider("https://github.com/owner/repo/pull/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "github" {
		t.Errorf("name = %q, want %q", name, "github")
	}
}

func TestDetectProviderUnknownHost(t *testing.T) {
	_, err := detectProvider("https://gitlab.com/owner/repo/-/merge_requests/1")
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestDetectProviderInvalidURL(t *testing.T) {
	_, err := detectProvider("not-a-url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestResolveGitHub(t *testing.T) {
	reg := NewRegistry("test-token")
	p, err := reg.Resolve("github", "https://github.com/o/r/pull/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("provider is nil")
	}
}

func TestResolveAutoDetectGitHub(t *testing.T) {
	reg := NewRegistry("test-token")
	p, err := reg.Resolve("", "https://github.com/o/r/pull/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("provider is nil")
	}
}

func TestResolveAutoDetectFails(t *testing.T) {
	reg := NewRegistry("")
	_, err := reg.Resolve("", "https://unknown.example.com/pr/1")
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
}

func TestResolvePluginNotFound(t *testing.T) {
	reg := NewRegistry("")
	_, err := reg.Resolve("nonexistent", "https://example.com/pr/1")
	if err == nil {
		t.Fatal("expected error for missing plugin binary")
	}
}

func TestNewRegistryWithGitHubBaseURL(t *testing.T) {
	reg := NewRegistryWithGitHubBaseURL("test-token", "http://localhost:9999")
	if reg.githubToken != "test-token" {
		t.Errorf("githubToken = %q", reg.githubToken)
	}
	if reg.githubBaseURL != "http://localhost:9999" {
		t.Errorf("githubBaseURL = %q", reg.githubBaseURL)
	}

	// Should still resolve to a github provider for github.com URLs.
	p, err := reg.Resolve("", "https://github.com/o/r/pull/1")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p == nil {
		t.Fatal("provider is nil")
	}
}
