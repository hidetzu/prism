package github_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider/github"
)

func TestProviderImplementsInterface(t *testing.T) {
	// Compile-time check is in github.go, but verify at runtime too.
	p := github.NewProvider("")
	ref, err := p.Parse("https://github.com/owner/repo/pull/1")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if ref.Provider != "github" {
		t.Errorf("Provider = %q, want %q", ref.Provider, "github")
	}
	_ = domain.PRRef(ref) // type assertion
}
