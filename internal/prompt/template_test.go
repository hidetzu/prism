package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/prompt"
)

func TestRenderFromTemplate(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "review.tmpl")
	content := `Review: {{.PR.Title}}
Type: {{.Analysis.ChangeType}}
Mode: {{.Mode}}
Lang: {{.Lang}}
Files:
{{range .PR.ChangedFiles}}- {{.Path}}
{{end}}`
	if err := os.WriteFile(tmplPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	bundle, err := prompt.RenderFromTemplate(tmplPath, domain.PromptModeLight, samplePR(), sampleResult(), "en")
	if err != nil {
		t.Fatalf("RenderFromTemplate: %v", err)
	}

	if bundle.Mode != domain.PromptModeLight {
		t.Errorf("Mode = %q, want %q", bundle.Mode, domain.PromptModeLight)
	}
	if bundle.SystemPrompt == "" {
		t.Error("SystemPrompt is empty")
	}
	if !strings.Contains(bundle.UserPrompt, "Add OAuth2 login") {
		t.Error("UserPrompt missing PR title")
	}
	if !strings.Contains(bundle.UserPrompt, "feature") {
		t.Error("UserPrompt missing change type")
	}
	if !strings.Contains(bundle.UserPrompt, "internal/auth/oauth.go") {
		t.Error("UserPrompt missing file path")
	}
}

func TestRenderFromTemplateJA(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "review.tmpl")
	content := `Lang: {{.Lang}}`
	if err := os.WriteFile(tmplPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	bundle, err := prompt.RenderFromTemplate(tmplPath, domain.PromptModeLight, samplePR(), sampleResult(), "ja")
	if err != nil {
		t.Fatalf("RenderFromTemplate: %v", err)
	}
	if !strings.Contains(bundle.UserPrompt, "Lang: ja") {
		t.Errorf("UserPrompt = %q, expected Lang: ja", bundle.UserPrompt)
	}
}

func TestRenderFromTemplateMissingFile(t *testing.T) {
	_, err := prompt.RenderFromTemplate("/nonexistent/template.tmpl", domain.PromptModeLight, samplePR(), sampleResult(), "en")
	if err == nil {
		t.Fatal("expected error for missing template file")
	}
}

func TestRenderFromTemplateInvalidSyntax(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "bad.tmpl")
	if err := os.WriteFile(tmplPath, []byte("{{.Invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := prompt.RenderFromTemplate(tmplPath, domain.PromptModeLight, samplePR(), sampleResult(), "en")
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}
