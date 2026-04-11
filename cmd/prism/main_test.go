package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hidetzu/prism/internal/domain"
)

func TestRootCommand(t *testing.T) {
	cmd := rootCmd()
	if cmd.Use != "prism" {
		t.Errorf("Use = %q, want %q", cmd.Use, "prism")
	}
	if cmd.Version != version {
		t.Errorf("Version = %q, want %q", cmd.Version, version)
	}
}

func TestSubcommands(t *testing.T) {
	cmd := rootCmd()

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}

	for _, want := range []string{"analyze <PR_URL>", "prompt <PR_URL>", "fetch <PR_URL>"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestAnalyzeFlags(t *testing.T) {
	cmd := analyzeCmd()
	flags := []string{"format", "config"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("analyze missing flag --%s", name)
		}
	}
}

func TestPromptFlags(t *testing.T) {
	cmd := promptCmd()
	flags := []string{"mode", "format", "lang", "template", "config"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("prompt missing flag --%s", name)
		}
	}
}

func TestFetchFlags(t *testing.T) {
	cmd := fetchCmd()
	flags := []string{"format", "config"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("fetch missing flag --%s", name)
		}
	}
}

func TestExitCodeInvalidArgs(t *testing.T) {
	err := errors.New("wrapped: " + domain.ErrInvalidArgs.Error())
	wrapped := fmt.Errorf("%w: bad input", domain.ErrInvalidArgs)
	if exitCode(wrapped) != ExitInvalidArgs {
		t.Errorf("exitCode for ErrInvalidArgs = %d, want %d", exitCode(wrapped), ExitInvalidArgs)
	}
	_ = err
}

func TestExitCodeProviderError(t *testing.T) {
	wrapped := fmt.Errorf("%w: API failure", domain.ErrProvider)
	if exitCode(wrapped) != ExitProviderError {
		t.Errorf("exitCode for ErrProvider = %d, want %d", exitCode(wrapped), ExitProviderError)
	}
}

func TestExitCodeGeneral(t *testing.T) {
	err := errors.New("unknown error")
	if exitCode(err) != ExitGeneralError {
		t.Errorf("exitCode for unknown = %d, want %d", exitCode(err), ExitGeneralError)
	}
}

func TestAnalyzeRequiresURL(t *testing.T) {
	cmd := analyzeCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPromptRequiresURL(t *testing.T) {
	cmd := promptCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchRequiresURL(t *testing.T) {
	cmd := fetchCmd()
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderFlag(t *testing.T) {
	cmd := rootCmd()
	if cmd.PersistentFlags().Lookup("provider") == nil {
		t.Error("missing persistent flag --provider")
	}
}
