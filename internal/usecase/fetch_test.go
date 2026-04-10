package usecase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/usecase"
)

func TestFetchJSON(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Fetch(context.Background(), p, sampleRef(), usecase.FetchOptions{
		Format: "json",
	}, &buf)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	var pr domain.PullRequest
	if err := json.Unmarshal(buf.Bytes(), &pr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pr.Title != "Fix null pointer in handler" {
		t.Errorf("title = %q", pr.Title)
	}
}

func TestFetchText(t *testing.T) {
	p := &mockProvider{pr: samplePR()}
	var buf bytes.Buffer

	err := usecase.Fetch(context.Background(), p, sampleRef(), usecase.FetchOptions{
		Format: "text",
	}, &buf)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "owner/repo") {
		t.Error("output missing repository")
	}
	if !strings.Contains(out, "internal/handler.go") {
		t.Error("output missing file path")
	}
}

func TestFetchProviderError(t *testing.T) {
	p := &mockProvider{err: errMock}
	var buf bytes.Buffer

	err := usecase.Fetch(context.Background(), p, sampleRef(), usecase.FetchOptions{}, &buf)
	if err == nil {
		t.Fatal("expected error")
	}
}
