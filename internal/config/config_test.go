package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hidetzu/prism/internal/config"
)

func TestLoadMissingFileReturnsZero(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultFormat != "" {
		t.Errorf("DefaultFormat = %q, want empty", cfg.DefaultFormat)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
github_token: file-token
default_format: markdown
default_mode: detailed
default_lang: ja
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear env to avoid interference.
	t.Setenv("GITHUB_TOKEN", "")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GitHubToken != "file-token" {
		t.Errorf("GitHubToken = %q, want %q", cfg.GitHubToken, "file-token")
	}
	if cfg.DefaultFormat != "markdown" {
		t.Errorf("DefaultFormat = %q, want %q", cfg.DefaultFormat, "markdown")
	}
	if cfg.DefaultMode != "detailed" {
		t.Errorf("DefaultMode = %q, want %q", cfg.DefaultMode, "detailed")
	}
	if cfg.DefaultLang != "ja" {
		t.Errorf("DefaultLang = %q, want %q", cfg.DefaultLang, "ja")
	}
}

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `github_token: file-token`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GITHUB_TOKEN", "env-token")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GitHubToken != "env-token" {
		t.Errorf("GitHubToken = %q, want %q (env should override file)", cfg.GitHubToken, "env-token")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadPRISMConfigEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	content := `default_format: text`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PRISM_CONFIG", path)
	t.Setenv("GITHUB_TOKEN", "")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DefaultFormat != "text" {
		t.Errorf("DefaultFormat = %q, want %q", cfg.DefaultFormat, "text")
	}
}
