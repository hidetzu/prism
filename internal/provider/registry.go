package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider/github"
	"github.com/hidetzu/prism/internal/provider/plugin"
)

// Registry manages built-in and plugin providers.
type Registry struct {
	githubToken   string
	githubBaseURL string // empty = use the real GitHub API
}

// NewRegistry creates a provider registry.
func NewRegistry(githubToken string) *Registry {
	return &Registry{githubToken: githubToken}
}

// NewRegistryWithGitHubBaseURL creates a registry that points the GitHub
// provider at a custom base URL. Used by tests to redirect API calls to a
// local httptest server.
func NewRegistryWithGitHubBaseURL(githubToken, baseURL string) *Registry {
	return &Registry{githubToken: githubToken, githubBaseURL: baseURL}
}

// Resolve returns a Provider for the given PR URL.
// If providerName is non-empty, it uses that provider directly.
// If providerName is empty, it auto-detects the provider from the URL.
func (r *Registry) Resolve(providerName string, prURL string) (Provider, error) {
	if providerName == "" {
		detected, err := detectProvider(prURL)
		if err != nil {
			return nil, err
		}
		providerName = detected
	}

	switch providerName {
	case "github":
		if r.githubBaseURL != "" {
			return github.NewProviderWithClient(http.DefaultClient, r.githubToken, r.githubBaseURL), nil
		}
		return github.NewProvider(r.githubToken), nil
	default:
		return r.resolvePlugin(providerName, prURL)
	}
}

func detectProvider(prURL string) (string, error) {
	u, err := url.Parse(prURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf(
			"cannot auto-detect provider from %q; use --provider to specify one", prURL,
		)
	}

	switch u.Host {
	case "github.com":
		return "github", nil
	default:
		return "", fmt.Errorf(
			"cannot auto-detect provider for host %q; use --provider to specify one", u.Host,
		)
	}
}

func (r *Registry) resolvePlugin(name string, prURL string) (Provider, error) {
	binary := "prism-provider-" + name
	path, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: provider %q not found: %s is not on PATH", domain.ErrProvider, name, binary,
		)
	}
	return plugin.NewProvider(name, path, prURL), nil
}
