package usecase_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/usecase"
)

func TestPromptLight(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode:   "light",
		Format: "text",
	}, &buf)
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Fix null pointer in handler") {
		t.Error("output missing PR title")
	}
	if strings.Contains(out, "```diff") {
		t.Error("light mode should not contain diffs")
	}
}

func TestPromptDetailed(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode:   "detailed",
		Format: "text",
	}, &buf)
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "```diff") {
		t.Error("detailed mode should contain diffs")
	}
}

func TestPromptJSON(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode:   "light",
		Format: "json",
	}, &buf)
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	if !strings.Contains(buf.String(), `"Mode"`) {
		t.Error("JSON output missing Mode field")
	}
}

func TestPromptCross(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode:   "cross",
		Format: "text",
		Lang:   "en",
	}, &buf)
	if err != nil {
		t.Fatalf("Prompt cross: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Module Structure") {
		t.Error("cross output missing module structure")
	}
	if !strings.Contains(out, "Cross-File Relationships") {
		t.Error("cross output missing relationships")
	}
}

func TestPromptJA(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode:   "light",
		Format: "text",
		Lang:   "ja",
	}, &buf)
	if err != nil {
		t.Fatalf("Prompt ja: %v", err)
	}

	if !strings.Contains(buf.String(), "コードレビュアー") {
		t.Error("ja prompt missing Japanese text")
	}
}

func TestPromptInvalidMode(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{
		Mode: "invalid",
	}, &buf)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestPromptProviderError(t *testing.T) {
	p := &mockProvider{err: errMock}
	var buf bytes.Buffer

	err := usecase.Prompt(context.Background(), p, sampleRef(), usecase.PromptOptions{}, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}
