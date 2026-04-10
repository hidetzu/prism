package formatter

import (
	"encoding/json"
	"io"

	"github.com/hidetzu/prism/internal/domain"
)

// FormatJSON writes the analysis result as JSON to w.
func FormatJSON(w io.Writer, providerName string, pr domain.PullRequest, result domain.AnalysisResult) error {
	out := toOutput(providerName, pr, result)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

func toOutput(providerName string, pr domain.PullRequest, result domain.AnalysisResult) Output {
	files := make([]ChangedFileOutput, 0, len(pr.ChangedFiles))
	for _, f := range pr.ChangedFiles {
		files = append(files, ChangedFileOutput{
			Path:        f.Path,
			Status:      string(f.Status),
			Additions:   f.Additions,
			Deletions:   f.Deletions,
			Language:    f.Language,
			IsTest:      f.IsTest,
			IsConfig:    f.IsConfig,
			IsGenerated: f.IsGenerated,
		})
	}

	axes := make([]string, 0, len(result.ReviewAxes))
	for _, a := range result.ReviewAxes {
		axes = append(axes, string(a))
	}

	return Output{
		Provider: providerName,
		PullRequest: PullRequestOutput{
			Repository:   pr.Repository,
			ID:           pr.ID,
			Title:        pr.Title,
			Author:       pr.Author,
			SourceBranch: pr.SourceBranch,
			TargetBranch: pr.TargetBranch,
			Description:  pr.Description,
		},
		ChangedFiles: files,
		Analysis: AnalysisOutput{
			ChangeType:    string(result.ChangeType),
			RiskLevel:     string(result.RiskLevel),
			AffectedAreas: nonNilSlice(result.AffectedAreas),
			ReviewAxes:    axes,
			RelatedFiles:  nonNilSlice(result.RelatedFiles),
			Warnings:      nonNilSlice(result.Warnings),
			Summary:       result.Summary,
		},
	}
}

// nonNilSlice ensures nil slices become empty arrays in JSON.
func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
