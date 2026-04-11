package classifier

import (
	"strings"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/fileutil"
)

// Classify determines the change type of a pull request based on its metadata and files.
func Classify(pr domain.PullRequest) domain.ChangeType {
	if ct, ok := classifyByFiles(pr.ChangedFiles); ok {
		return ct
	}
	if ct, ok := classifyByTitle(pr.Title); ok {
		return ct
	}
	if ct, ok := classifyByDescription(pr.Description); ok {
		return ct
	}
	return domain.ChangeTypeFeature
}

func classifyByFiles(files []domain.ChangedFile) (domain.ChangeType, bool) {
	if len(files) == 0 {
		return "", false
	}

	allTest := true
	allDocs := true
	allConfig := true
	allInfra := true
	allDeps := true

	for _, f := range files {
		// IsTest/IsConfig flags are set by provider via fileutil.
		// Path-based fallback ensures correctness when flags are not pre-populated.
		if !f.IsTest && !fileutil.IsTestFile(f.Path) {
			allTest = false
		}
		if !fileutil.IsDocFile(f.Path) {
			allDocs = false
		}
		if !fileutil.IsInfraFile(f.Path) {
			allInfra = false
		}
		if (!f.IsConfig && !fileutil.IsConfigFile(f.Path)) || fileutil.IsInfraFile(f.Path) {
			allConfig = false
		}
		if !fileutil.IsDependencyFile(f.Path) {
			allDeps = false
		}
	}

	switch {
	case allTest:
		return domain.ChangeTypeTestOnly, true
	case allDocs:
		return domain.ChangeTypeDocsOnly, true
	case allDeps:
		return domain.ChangeTypeDependencyUpdate, true
	case allInfra:
		return domain.ChangeTypeInfraChange, true
	case allConfig:
		return domain.ChangeTypeConfigChange, true
	}

	return "", false
}

func classifyByTitle(title string) (domain.ChangeType, bool) {
	lower := strings.ToLower(title)

	bugKeywords := []string{"fix", "bug", "hotfix", "patch", "resolve", "issue"}
	for _, kw := range bugKeywords {
		if strings.Contains(lower, kw) {
			return domain.ChangeTypeBugfix, true
		}
	}

	refactorKeywords := []string{"refactor", "restructure", "reorganize", "cleanup", "clean up"}
	for _, kw := range refactorKeywords {
		if strings.Contains(lower, kw) {
			return domain.ChangeTypeRefactor, true
		}
	}

	depKeywords := []string{"bump", "upgrade", "update dependency", "update dependencies", "renovate", "dependabot"}
	for _, kw := range depKeywords {
		if strings.Contains(lower, kw) {
			return domain.ChangeTypeDependencyUpdate, true
		}
	}

	return "", false
}

func classifyByDescription(desc string) (domain.ChangeType, bool) {
	lower := strings.ToLower(desc)

	if strings.Contains(lower, "fixes #") || strings.Contains(lower, "closes #") || strings.Contains(lower, "resolves #") {
		return domain.ChangeTypeBugfix, true
	}

	return "", false
}
