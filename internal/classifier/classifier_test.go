package classifier_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/classifier"
	"github.com/hidetzu/prism/internal/domain"
)

func TestClassifyTestOnly(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add unit tests",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/handler_test.go", IsTest: true},
			{Path: "tests/integration_test.go", IsTest: true},
		},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeTestOnly {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeTestOnly)
	}
}

func TestClassifyDocsOnly(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Update documentation",
		ChangedFiles: []domain.ChangedFile{
			{Path: "README.md"},
			{Path: "docs/architecture.md"},
		},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeDocsOnly {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeDocsOnly)
	}
}

func TestClassifyDependencyUpdate(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Update Go modules",
		ChangedFiles: []domain.ChangedFile{
			{Path: "go.mod"},
			{Path: "go.sum"},
		},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeDependencyUpdate {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeDependencyUpdate)
	}
}

func TestClassifyInfraChange(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Update CI pipeline",
		ChangedFiles: []domain.ChangedFile{
			{Path: ".github/workflows/ci.yml"},
		},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeInfraChange {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeInfraChange)
	}
}

func TestClassifyBugfixByTitle(t *testing.T) {
	tests := []struct {
		title string
	}{
		{"Fix null pointer in handler"},
		{"hotfix: login crash"},
		{"Bug: incorrect calculation"},
		{"Resolve timeout issue"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			pr := domain.PullRequest{
				Title:        tt.title,
				ChangedFiles: []domain.ChangedFile{{Path: "internal/handler.go"}},
			}
			got := classifier.Classify(pr)
			if got != domain.ChangeTypeBugfix {
				t.Errorf("got %q, want %q for title %q", got, domain.ChangeTypeBugfix, tt.title)
			}
		})
	}
}

func TestClassifyRefactorByTitle(t *testing.T) {
	pr := domain.PullRequest{
		Title:        "Refactor auth middleware",
		ChangedFiles: []domain.ChangedFile{{Path: "internal/middleware.go"}},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeRefactor {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeRefactor)
	}
}

func TestClassifyBugfixByDescription(t *testing.T) {
	pr := domain.PullRequest{
		Title:        "Update handler logic",
		Description:  "Fixes #123 — the handler was not checking for nil",
		ChangedFiles: []domain.ChangedFile{{Path: "internal/handler.go"}},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeBugfix {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeBugfix)
	}
}

func TestClassifyFeatureDefault(t *testing.T) {
	pr := domain.PullRequest{
		Title:        "Add OAuth2 login",
		Description:  "Implements OAuth2 flow",
		ChangedFiles: []domain.ChangedFile{{Path: "internal/auth/oauth.go"}},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeFeature {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeFeature)
	}
}

func TestClassifyConfigChange(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Update ESLint config",
		ChangedFiles: []domain.ChangedFile{
			{Path: ".eslintrc.json", IsConfig: true},
		},
	}
	got := classifier.Classify(pr)
	if got != domain.ChangeTypeConfigChange {
		t.Errorf("got %q, want %q", got, domain.ChangeTypeConfigChange)
	}
}
