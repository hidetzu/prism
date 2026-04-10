package domain

import "errors"

// Sentinel errors for categorizing error types.
// These are used by the CLI to determine the correct exit code.
var (
	// ErrInvalidArgs indicates invalid user input (bad URL, bad flag values).
	ErrInvalidArgs = errors.New("invalid arguments")

	// ErrProvider indicates a provider-level failure (API error, auth failure).
	ErrProvider = errors.New("provider error")

	// ErrAnalysis indicates a failure during analysis or formatting.
	ErrAnalysis = errors.New("analysis error")
)
