package analyzer_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/analyzer"
	"github.com/hidetzu/prism/internal/domain"
)

func TestRiskLevelLowForDocsOnly(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "README.md", Additions: 5},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeDocsOnly)
	if result.RiskLevel != domain.RiskLevelLow {
		t.Errorf("RiskLevel = %q, want %q", result.RiskLevel, domain.RiskLevelLow)
	}
}

func TestRiskLevelHighForLargeChange(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/big.go", Additions: 400, Deletions: 200},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	if result.RiskLevel != domain.RiskLevelHigh {
		t.Errorf("RiskLevel = %q, want %q", result.RiskLevel, domain.RiskLevelHigh)
	}
}

func TestRiskLevelHighForSecurityFile(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/token.go", Additions: 10},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	if result.RiskLevel != domain.RiskLevelHigh {
		t.Errorf("RiskLevel = %q, want %q", result.RiskLevel, domain.RiskLevelHigh)
	}
}

func TestRiskLevelMediumForConfig(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "config.yaml", IsConfig: true, Additions: 5},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeConfigChange)
	if result.RiskLevel != domain.RiskLevelMedium {
		t.Errorf("RiskLevel = %q, want %q", result.RiskLevel, domain.RiskLevelMedium)
	}
}

func TestReviewAxesForFeature(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/handler.go", Additions: 50},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	axesMap := make(map[domain.ReviewAxis]bool)
	for _, a := range result.ReviewAxes {
		axesMap[a] = true
	}
	if !axesMap[domain.ReviewAxisTestCoverage] {
		t.Error("expected test coverage axis for feature")
	}
	if !axesMap[domain.ReviewAxisEdgeCases] {
		t.Error("expected edge cases axis for feature")
	}
}

func TestReviewAxesForSecurityFile(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/password.go", Additions: 30},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	axesMap := make(map[domain.ReviewAxis]bool)
	for _, a := range result.ReviewAxes {
		axesMap[a] = true
	}
	if !axesMap[domain.ReviewAxisSecurity] {
		t.Error("expected security axis for auth file")
	}
}

func TestWarningNoTests(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add feature",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/handler.go", Additions: 50},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	found := false
	for _, w := range result.Warnings {
		if w == "No test files included in this change" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'No test files' warning")
	}
}

func TestNoWarningWhenTestsPresent(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add feature",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/handler.go", Additions: 50},
			{Path: "internal/handler_test.go", Additions: 30, IsTest: true},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	for _, w := range result.Warnings {
		if w == "No test files included in this change" {
			t.Error("unexpected 'No test files' warning when tests are present")
		}
	}
}

func TestAffectedAreas(t *testing.T) {
	pr := domain.PullRequest{
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/handler.go"},
			{Path: "internal/billing/payment.go"},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	areas := make(map[string]bool)
	for _, a := range result.AffectedAreas {
		areas[a] = true
	}
	if !areas["auth"] {
		t.Error("expected 'auth' in affected areas")
	}
	if !areas["billing"] {
		t.Error("expected 'billing' in affected areas")
	}
}

func TestSummaryFormat(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add login",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth.go", Additions: 20, Deletions: 5},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	if result.Summary == "" {
		t.Error("summary should not be empty")
	}
}

func TestWarningLargeChangeSet(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Big refactor",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/big.go", Additions: 400, Deletions: 200},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeRefactor)
	found := false
	for _, w := range result.Warnings {
		if w == "Large change set — consider splitting into smaller PRs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected large change set warning")
	}
}

func TestRelatedFilesGoTestPair(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add feature",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/handler.go", Additions: 30},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	related := make(map[string]bool)
	for _, r := range result.RelatedFiles {
		related[r] = true
	}
	if !related["internal/auth/handler_test.go"] {
		t.Error("expected handler_test.go in related files")
	}
}

func TestRelatedFilesTestToSource(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add tests",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/handler_test.go", IsTest: true, Additions: 50},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeTestOnly)
	related := make(map[string]bool)
	for _, r := range result.RelatedFiles {
		related[r] = true
	}
	if !related["internal/auth/handler.go"] {
		t.Error("expected handler.go in related files")
	}
}

func TestRelatedFilesExcludesChanged(t *testing.T) {
	pr := domain.PullRequest{
		Title: "Add feature with test",
		ChangedFiles: []domain.ChangedFile{
			{Path: "internal/auth/handler.go", Additions: 30},
			{Path: "internal/auth/handler_test.go", IsTest: true, Additions: 20},
		},
	}
	result := analyzer.Analyze(pr, domain.ChangeTypeFeature)
	for _, r := range result.RelatedFiles {
		if r == "internal/auth/handler.go" || r == "internal/auth/handler_test.go" {
			t.Errorf("related files should not include changed file %q", r)
		}
	}
}
