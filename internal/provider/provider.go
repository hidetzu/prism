package provider

import (
	"context"

	"github.com/hidetzu/prism/internal/domain"
)

// Provider abstracts the source of pull request data.
// Each PR hosting service (GitHub, CodeCommit, etc.) implements this interface.
type Provider interface {
	// Parse extracts a PRRef from a user-supplied input string (typically a URL).
	Parse(input string) (domain.PRRef, error)

	// FetchPullRequest retrieves pull request data for the given reference.
	FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error)
}
