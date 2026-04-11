package provider

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/provider/github"
	"github.com/hidetzu/prism/internal/provider/plugin"
)

// Registry manages built-in and plugin providers.
type Registry struct {
	githubToken string
}

// NewRegistry creates a provider registry.
func NewRegistry(githubToken string) *Registry {
	return &Registry{githubToken: githubToken}
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
