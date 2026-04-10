package analyzer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hidetzu/prism/internal/domain"
)

// Analyze estimates risk, suggests review axes, and identifies affected areas and warnings.
func Analyze(pr domain.PullRequest, changeType domain.ChangeType) domain.AnalysisResult {
	areas := detectAffectedAreas(pr.ChangedFiles)
	risk := estimateRisk(pr, changeType)
	axes := suggestReviewAxes(pr, changeType)
	warnings := detectWarnings(pr, changeType)
	summary := buildSummary(pr, changeType)

	related := detectRelatedFiles(pr.ChangedFiles)

	return domain.AnalysisResult{
		ChangeType:    changeType,
		RiskLevel:     risk,
		AffectedAreas: areas,
		ReviewAxes:    axes,
		RelatedFiles:  related,
		Warnings:      warnings,
		Summary:       summary,
	}
}

func estimateRisk(pr domain.PullRequest, changeType domain.ChangeType) domain.RiskLevel {
	// Low-risk change types.
	switch changeType {
	case domain.ChangeTypeTestOnly, domain.ChangeTypeDocsOnly:
		return domain.RiskLevelLow
	}

	totalChanges := 0
	for _, f := range pr.ChangedFiles {
		totalChanges += f.Additions + f.Deletions
	}

	fileCount := len(pr.ChangedFiles)

	// High risk indicators.
	if totalChanges > 500 || fileCount > 20 {
		return domain.RiskLevelHigh
	}

	hasSecurityFile := false
	hasConfigFile := false
	for _, f := range pr.ChangedFiles {
		if f.IsConfig {
			hasConfigFile = true
		}
		lower := strings.ToLower(f.Path)
		if strings.Contains(lower, "auth") || strings.Contains(lower, "security") ||
			strings.Contains(lower, "crypto") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "credential") || strings.Contains(lower, "password") {
			hasSecurityFile = true
		}
	}

	if hasSecurityFile {
		return domain.RiskLevelHigh
	}

	if totalChanges > 200 || fileCount > 10 || hasConfigFile {
		return domain.RiskLevelMedium
	}

	return domain.RiskLevelLow
}

func suggestReviewAxes(pr domain.PullRequest, changeType domain.ChangeType) []domain.ReviewAxis {
	axes := make(map[domain.ReviewAxis]bool)

	// Always suggest based on change type.
	switch changeType {
	case domain.ChangeTypeFeature:
		axes[domain.ReviewAxisTestCoverage] = true
		axes[domain.ReviewAxisEdgeCases] = true
	case domain.ChangeTypeBugfix:
		axes[domain.ReviewAxisEdgeCases] = true
		axes[domain.ReviewAxisErrorHandling] = true
	case domain.ChangeTypeRefactor:
		axes[domain.ReviewAxisReadability] = true
		axes[domain.ReviewAxisSeparationOfConcern] = true
		axes[domain.ReviewAxisBackwardCompat] = true
	case domain.ChangeTypeConfigChange:
		axes[domain.ReviewAxisConfigSafety] = true
	case domain.ChangeTypeDependencyUpdate:
		axes[domain.ReviewAxisBackwardCompat] = true
		axes[domain.ReviewAxisSecurity] = true
	case domain.ChangeTypeInfraChange:
		axes[domain.ReviewAxisConfigSafety] = true
		axes[domain.ReviewAxisSecurity] = true
	}

	// File-based axes.
	for _, f := range pr.ChangedFiles {
		lower := strings.ToLower(f.Path)
		if strings.Contains(lower, "auth") || strings.Contains(lower, "security") ||
			strings.Contains(lower, "crypto") || strings.Contains(lower, "password") {
			axes[domain.ReviewAxisSecurity] = true
		}
		if strings.Contains(lower, "api") || strings.Contains(lower, "handler") ||
			strings.Contains(lower, "endpoint") || strings.Contains(lower, "route") {
			axes[domain.ReviewAxisErrorHandling] = true
			axes[domain.ReviewAxisBackwardCompat] = true
		}
		if strings.Contains(lower, "query") || strings.Contains(lower, "db") ||
			strings.Contains(lower, "cache") || strings.Contains(lower, "index") {
			axes[domain.ReviewAxisPerformance] = true
		}
	}

	// Large changes deserve readability review.
	totalChanges := 0
	for _, f := range pr.ChangedFiles {
		totalChanges += f.Additions + f.Deletions
	}
	if totalChanges > 300 {
		axes[domain.ReviewAxisReadability] = true
	}

	result := make([]domain.ReviewAxis, 0, len(axes))
	// Return in a stable order.
	for _, axis := range allAxes {
		if axes[axis] {
			result = append(result, axis)
		}
	}
	return result
}

var allAxes = []domain.ReviewAxis{
	domain.ReviewAxisErrorHandling,
	domain.ReviewAxisBackwardCompat,
	domain.ReviewAxisTestCoverage,
	domain.ReviewAxisPerformance,
	domain.ReviewAxisSecurity,
	domain.ReviewAxisConfigSafety,
	domain.ReviewAxisEdgeCases,
	domain.ReviewAxisReadability,
	domain.ReviewAxisSeparationOfConcern,
}

func detectAffectedAreas(files []domain.ChangedFile) []string {
	areas := make(map[string]bool)
	for _, f := range files {
		dir := filepath.Dir(f.Path)
		parts := strings.Split(dir, "/")
		// Use the first meaningful directory as the area.
		for _, p := range parts {
			if p == "." || p == "internal" || p == "src" || p == "lib" || p == "pkg" || p == "app" {
				continue
			}
			if p != "" {
				areas[p] = true
				break
			}
		}
	}

	result := make([]string, 0, len(areas))
	for area := range areas {
		result = append(result, area)
	}
	sort.Strings(result)
	return result
}

func detectWarnings(pr domain.PullRequest, changeType domain.ChangeType) []string {
	var warnings []string

	hasTest := false
	hasSource := false
	for _, f := range pr.ChangedFiles {
		if f.IsTest {
			hasTest = true
		} else if !f.IsConfig && !f.IsGenerated {
			hasSource = true
		}
	}

	if hasSource && !hasTest && changeType != domain.ChangeTypeDocsOnly && changeType != domain.ChangeTypeConfigChange {
		warnings = append(warnings, "No test files included in this change")
	}

	totalChanges := 0
	for _, f := range pr.ChangedFiles {
		totalChanges += f.Additions + f.Deletions
	}
	if totalChanges > 500 {
		warnings = append(warnings, "Large change set — consider splitting into smaller PRs")
	}

	if len(pr.ChangedFiles) > 20 {
		warnings = append(warnings, "High number of changed files — review may require extra attention")
	}

	generatedCount := 0
	for _, f := range pr.ChangedFiles {
		if f.IsGenerated {
			generatedCount++
		}
	}
	if generatedCount > 0 {
		warnings = append(warnings, "Contains generated files — verify regeneration is intentional")
	}

	return warnings
}

func buildSummary(pr domain.PullRequest, changeType domain.ChangeType) string {
	fileCount := len(pr.ChangedFiles)
	totalAdd := 0
	totalDel := 0
	for _, f := range pr.ChangedFiles {
		totalAdd += f.Additions
		totalDel += f.Deletions
	}

	filesWord := "files changed"
	if fileCount == 1 {
		filesWord = "file changed"
	}

	return fmt.Sprintf("%s: %s (%d %s, +%d/-%d)",
		changeType, pr.Title, fileCount, filesWord, totalAdd, totalDel)
}

// detectRelatedFiles infers files that are likely related to the changed files
// but not themselves changed. This uses naming conventions and path patterns.
func detectRelatedFiles(files []domain.ChangedFile) []string {
	changed := make(map[string]bool)
	for _, f := range files {
		changed[f.Path] = true
	}

	candidates := make(map[string]bool)
	for _, f := range files {
		for _, rel := range relatedCandidates(f.Path) {
			if !changed[rel] {
				candidates[rel] = true
			}
		}
	}

	result := make([]string, 0, len(candidates))
	for c := range candidates {
		result = append(result, c)
	}
	sort.Strings(result)
	return result
}

// relatedCandidates returns potential related file paths for a given file.
func relatedCandidates(path string) []string {
	var candidates []string

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	nameNoExt := strings.TrimSuffix(base, ext)

	// Go: source ↔ test pair
	if strings.HasSuffix(base, "_test.go") {
		source := strings.TrimSuffix(base, "_test.go") + ".go"
		candidates = append(candidates, filepath.Join(dir, source))
	} else if ext == ".go" {
		candidates = append(candidates, filepath.Join(dir, nameNoExt+"_test.go"))
	}

	// JS/TS: source ↔ test pair
	for _, testSuffix := range []string{".test", ".spec"} {
		for _, jsExt := range []string{".js", ".ts", ".tsx", ".jsx"} {
			if strings.HasSuffix(nameNoExt, testSuffix) && ext == jsExt {
				source := strings.TrimSuffix(nameNoExt, testSuffix) + ext
				candidates = append(candidates, filepath.Join(dir, source))
			} else if ext == jsExt {
				candidates = append(candidates, filepath.Join(dir, nameNoExt+testSuffix+ext))
			}
		}
	}

	// Interface/implementation: if handler.go changed, suggest service.go, repository.go in same dir
	if ext == ".go" && !strings.HasSuffix(base, "_test.go") {
		for _, peer := range []string{"handler", "service", "repository", "controller", "middleware", "model", "store"} {
			peerFile := filepath.Join(dir, peer+".go")
			if peerFile != path {
				candidates = append(candidates, peerFile)
			}
		}
	}

	return candidates
}
