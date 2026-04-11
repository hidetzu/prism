package usecase_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/usecase"
)

func TestValidatePromptOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    usecase.PromptOptions
		wantErr bool
	}{
		{"valid all", usecase.PromptOptions{Mode: "light", Format: "text", Lang: "en"}, false},
		{"valid detailed", usecase.PromptOptions{Mode: "detailed"}, false},
		{"valid cross", usecase.PromptOptions{Mode: "cross"}, false},
		{"valid ja", usecase.PromptOptions{Lang: "ja"}, false},
		{"invalid mode", usecase.PromptOptions{Mode: "invalid"}, true},
		{"invalid lang", usecase.PromptOptions{Lang: "fr"}, true},
		{"invalid format", usecase.PromptOptions{Format: "xml"}, true},
		{"empty all", usecase.PromptOptions{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := usecase.ValidatePromptOptions(tt.opts)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateFetchOptions(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"json", false},
		{"text", false},
		{"", false},
		{"markdown", true},
	}
	for _, tt := range tests {
		err := usecase.ValidateFetchOptions(usecase.FetchOptions{Format: tt.format})
		if tt.wantErr && err == nil {
			t.Errorf("format %q: expected error", tt.format)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("format %q: unexpected error: %v", tt.format, err)
		}
	}
}
