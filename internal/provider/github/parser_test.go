package github_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/provider/github"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOwner string
		wantRepo  string
		wantNum   int
		wantErr bool
	}{
		{
			name:      "valid PR URL",
			input:     "https://github.com/owner/repo/pull/123",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantNum:   123,
		},
		{
			name:      "valid PR URL with trailing slash",
			input:     "https://github.com/owner/repo/pull/456/",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantNum:   456,
		},
		{
			name:    "wrong host",
			input:   "https://gitlab.com/owner/repo/pull/1",
			wantErr: true,
		},
		{
			name:    "not a PR URL",
			input:   "https://github.com/owner/repo/issues/1",
			wantErr: true,
		},
		{
			name:    "missing PR number",
			input:   "https://github.com/owner/repo/pull",
			wantErr: true,
		},
		{
			name:    "non-numeric PR number",
			input:   "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "negative PR number",
			input:   "https://github.com/owner/repo/pull/-1",
			wantErr: true,
		},
		{
			name:    "zero PR number",
			input:   "https://github.com/owner/repo/pull/0",
			wantErr: true,
		},
		{
			name:    "repo URL without pull",
			input:   "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := github.Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Provider != "github" {
				t.Errorf("Provider = %q, want %q", ref.Provider, "github")
			}
			if ref.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", ref.Owner, tt.wantOwner)
			}
			if ref.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", ref.Repo, tt.wantRepo)
			}
			if ref.Number != tt.wantNum {
				t.Errorf("Number = %d, want %d", ref.Number, tt.wantNum)
			}
		})
	}
}
