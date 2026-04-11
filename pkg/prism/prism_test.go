package prism_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hidetzu/prism/pkg/prism"
)

func TestAnalyzeMissingURL(t *testing.T) {
	_, err := prism.Analyze(context.Background(), prism.AnalyzeOptions{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if !errors.Is(err, prism.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPromptMissingURL(t *testing.T) {
	_, err := prism.Prompt(context.Background(), prism.AnalyzeOptions{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if !errors.Is(err, prism.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPromptInvalidMode(t *testing.T) {
	_, err := prism.Prompt(context.Background(), prism.AnalyzeOptions{
		PRURL: "https://github.com/o/r/pull/1",
		Mode:  "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !errors.Is(err, prism.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPromptInvalidLanguage(t *testing.T) {
	_, err := prism.Prompt(context.Background(), prism.AnalyzeOptions{
		PRURL:    "https://github.com/o/r/pull/1",
		Language: "fr",
	})
	if err == nil {
		t.Fatal("expected error for invalid language")
	}
	if !errors.Is(err, prism.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAnalyzeAutoDetectFails(t *testing.T) {
	_, err := prism.Analyze(context.Background(), prism.AnalyzeOptions{
		PRURL: "https://unknown.example.com/pr/1",
	})
	if err == nil {
		t.Fatal("expected error for unknown host")
	}
	if !errors.Is(err, prism.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for auto-detect failure, got %v", err)
	}
}

func TestAnalyzeUnsupportedProvider(t *testing.T) {
	_, err := prism.Analyze(context.Background(), prism.AnalyzeOptions{
		PRURL:    "https://example.com/pr/1",
		Provider: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for missing plugin")
	}
	if !errors.Is(err, prism.ErrUnsupportedProvider) {
		t.Errorf("expected ErrUnsupportedProvider, got %v", err)
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	errs := []error{
		prism.ErrInvalidInput,
		prism.ErrUnsupportedProvider,
		prism.ErrAuthRequired,
		prism.ErrUpstreamFailure,
	}
	for i, e1 := range errs {
		for j, e2 := range errs {
			if i == j {
				continue
			}
			if errors.Is(e1, e2) {
				t.Errorf("errors[%d] should not match errors[%d]", i, j)
			}
		}
	}
}
