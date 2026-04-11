package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/fileutil"
)

const (
	// defaultTimeout is the maximum duration for a complete FetchPullRequest call.
	defaultTimeout = 30 * time.Second

	// maxPages limits pagination to prevent unbounded memory allocation.
	maxPages = 100
)

// HTTPClient abstracts HTTP requests for testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client fetches pull request data from the GitHub REST API.
type Client struct {
	httpClient HTTPClient
	token      string
	baseURL    string
}

// NewClient creates a GitHub API client.
// baseURL should be "https://api.github.com" for production.
func NewClient(httpClient HTTPClient, token string, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		token:      token,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

// FetchPullRequest retrieves a pull request and its changed files from GitHub.
func (c *Client) FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	pr, err := c.fetchPR(ctx, ref)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("fetch PR metadata: %w", err)
	}

	files, err := c.fetchFiles(ctx, ref)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("fetch PR files: %w", err)
	}

	pr.ChangedFiles = files
	return pr, nil
}

type ghPullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"base"`
}

type ghFile struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

func (c *Client) fetchPR(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, ref.Owner, ref.Repo, ref.Number)

	body, err := c.doGet(ctx, url)
	if err != nil {
		return domain.PullRequest{}, err
	}
	defer func() { _ = body.Close() }()

	var ghPR ghPullRequest
	if err := json.NewDecoder(body).Decode(&ghPR); err != nil {
		return domain.PullRequest{}, fmt.Errorf("decode PR response: %w", err)
	}

	return domain.PullRequest{
		Repository:   ghPR.Base.Repo.FullName,
		ID:           fmt.Sprintf("%d", ghPR.Number),
		Title:        ghPR.Title,
		Author:       ghPR.User.Login,
		SourceBranch: ghPR.Head.Ref,
		TargetBranch: ghPR.Base.Ref,
		Description:  ghPR.Body,
	}, nil
}

func (c *Client) fetchFiles(ctx context.Context, ref domain.PRRef) ([]domain.ChangedFile, error) {
	var allFiles []domain.ChangedFile

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files?per_page=100&page=%d",
			c.baseURL, ref.Owner, ref.Repo, ref.Number, page)

		body, err := c.doGet(ctx, url)
		if err != nil {
			return nil, err
		}

		var ghFiles []ghFile
		if err := json.NewDecoder(body).Decode(&ghFiles); err != nil {
			_ = body.Close()
			return nil, fmt.Errorf("decode files response: %w", err)
		}
		_ = body.Close()

		if len(ghFiles) == 0 {
			break
		}

		for _, f := range ghFiles {
			allFiles = append(allFiles, domain.ChangedFile{
				Path:        f.Filename,
				Status:      mapFileStatus(f.Status),
				Additions:   f.Additions,
				Deletions:   f.Deletions,
				Language:    fileutil.DetectLanguage(f.Filename),
				IsTest:      fileutil.IsTestFile(f.Filename),
				IsConfig:    fileutil.IsConfigFile(f.Filename),
				IsGenerated: fileutil.IsGeneratedFile(f.Filename),
				Patch:       f.Patch,
			})
		}

		if len(ghFiles) < 100 {
			break
		}
	}

	if len(allFiles) >= maxPages*100 {
		return nil, fmt.Errorf("PR has too many files (exceeded %d pages)", maxPages)
	}

	return allFiles, nil
}

func (c *Client) doGet(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, nil
}

func mapFileStatus(status string) domain.FileStatus {
	switch status {
	case "added":
		return domain.FileStatusAdded
	case "modified":
		return domain.FileStatusModified
	case "removed":
		return domain.FileStatusRemoved
	case "renamed":
		return domain.FileStatusRenamed
	default:
		return domain.FileStatusModified
	}
}
