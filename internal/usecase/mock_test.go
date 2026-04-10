package usecase_test

import (
	"context"
	"errors"

	"github.com/hidetzu/prism/internal/domain"
)

var errMock = errors.New("mock provider error")

type mockProvider struct {
	pr  domain.PullRequest
	err error
}

func (m *mockProvider) Parse(input string) (domain.PRRef, error) {
	return sampleRef(), nil
}

func (m *mockProvider) FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	if m.err != nil {
		return domain.PullRequest{}, m.err
	}
	return m.pr, nil
}

func sampleRef() domain.PRRef {
	return domain.PRRef{
		Provider: "github",
		Owner:    "owner",
		Repo:     "repo",
		Number:   42,
	}
}

func samplePR() domain.PullRequest {
	return domain.PullRequest{
		Repository:   "owner/repo",
		ID:           "42",
		Title:        "Fix null pointer in handler",
		Author:       "dev",
		SourceBranch: "fix/null-ptr",
		TargetBranch: "main",
		Description:  "Fixes #99",
		ChangedFiles: []domain.ChangedFile{
			{
				Path:      "internal/handler.go",
				Status:    domain.FileStatusModified,
				Additions: 10,
				Deletions: 2,
				Language:  "Go",
				Patch:     "@@ -1,5 +1,13 @@",
			},
		},
	}
}
