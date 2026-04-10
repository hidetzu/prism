package github

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hidetzu/prism/internal/domain"
)

const providerName = "github"

// Parse extracts a PRRef from a GitHub pull request URL.
// Expected format: https://github.com/{owner}/{repo}/pull/{number}
func Parse(input string) (domain.PRRef, error) {
	u, err := url.Parse(input)
	if err != nil {
		return domain.PRRef{}, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Host != "github.com" {
		return domain.PRRef{}, fmt.Errorf("unsupported host: %q, expected \"github.com\"", u.Host)
	}

	// path: /{owner}/{repo}/pull/{number}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return domain.PRRef{}, fmt.Errorf("invalid GitHub PR URL path: %q, expected /{owner}/{repo}/pull/{number}", u.Path)
	}

	number, err := strconv.Atoi(parts[3])
	if err != nil {
		return domain.PRRef{}, fmt.Errorf("invalid PR number %q: %w", parts[3], err)
	}

	if number <= 0 {
		return domain.PRRef{}, fmt.Errorf("PR number must be positive, got %d", number)
	}

	return domain.PRRef{
		Provider: providerName,
		Owner:    parts[0],
		Repo:     parts[1],
		Number:   number,
	}, nil
}
