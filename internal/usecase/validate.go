package usecase

import (
	"fmt"

	"github.com/hidetzu/prism/internal/domain"
)

var validAnalyzeFormats = map[string]bool{
	"json": true, "markdown": true, "text": true,
}

var validPromptModes = map[string]bool{
	"light": true, "detailed": true, "cross": true,
}

var validPromptFormats = map[string]bool{
	"text": true, "markdown": true, "json": true,
}

var validFetchFormats = map[string]bool{
	"json": true, "text": true,
}

var validLangs = map[string]bool{
	"en": true, "ja": true,
}

// ValidateAnalyzeOptions returns an error if options are invalid.
func ValidateAnalyzeOptions(opts AnalyzeOptions) error {
	if opts.Format != "" && !validAnalyzeFormats[opts.Format] {
		return fmt.Errorf("%w: invalid format %q: must be json, markdown, or text", domain.ErrInvalidArgs, opts.Format)
	}
	return nil
}

// ValidatePromptOptions returns an error if options are invalid.
func ValidatePromptOptions(opts PromptOptions) error {
	if opts.Mode != "" && !validPromptModes[opts.Mode] {
		return fmt.Errorf("%w: invalid mode %q: must be light, detailed, or cross", domain.ErrInvalidArgs, opts.Mode)
	}
	if opts.Format != "" && !validPromptFormats[opts.Format] {
		return fmt.Errorf("%w: invalid format %q: must be text, markdown, or json", domain.ErrInvalidArgs, opts.Format)
	}
	if opts.Lang != "" && !validLangs[opts.Lang] {
		return fmt.Errorf("%w: invalid language %q: must be en or ja", domain.ErrInvalidArgs, opts.Lang)
	}
	return nil
}

// ValidateFetchOptions returns an error if options are invalid.
func ValidateFetchOptions(opts FetchOptions) error {
	if opts.Format != "" && !validFetchFormats[opts.Format] {
		return fmt.Errorf("%w: invalid format %q: must be json or text", domain.ErrInvalidArgs, opts.Format)
	}
	return nil
}
