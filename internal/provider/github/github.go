package github

import (
	"context"
	"net/http"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider"
)

const apiBaseURL = "https://api.github.com"

// Provider implements the provider.Provider interface for GitHub.
type Provider struct {
	client *Client
}

// Verify interface compliance at compile time.
var _ provider.Provider = (*Provider)(nil)

// NewProvider creates a GitHub provider with the given token.
func NewProvider(token string) *Provider {
	return &Provider{
		client: NewClient(http.DefaultClient, token, apiBaseURL),
	}
}

// NewProviderWithClient creates a GitHub provider with a custom HTTP client and base URL.
// This is primarily used for testing.
func NewProviderWithClient(httpClient HTTPClient, token string, baseURL string) *Provider {
	return &Provider{
		client: NewClient(httpClient, token, baseURL),
	}
}

// Parse extracts a PRRef from a GitHub pull request URL.
func (p *Provider) Parse(input string) (domain.PRRef, error) {
	return Parse(input)
}

// FetchPullRequest retrieves pull request data from GitHub.
func (p *Provider) FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	return p.client.FetchPullRequest(ctx, ref)
}
