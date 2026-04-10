package formatter

// Output types with JSON tags matching docs/json-schema.md.

// Output is the top-level JSON output structure.
type Output struct {
	Provider    string             `json:"provider"`
	PullRequest PullRequestOutput  `json:"pull_request"`
	ChangedFiles []ChangedFileOutput `json:"changed_files"`
	Analysis    AnalysisOutput     `json:"analysis"`
}

// PullRequestOutput represents PR metadata in the output.
type PullRequestOutput struct {
	Repository   string `json:"repository"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Description  string `json:"description"`
}

// ChangedFileOutput represents a changed file in the output.
type ChangedFileOutput struct {
	Path        string `json:"path"`
	Status      string `json:"status"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Language    string `json:"language"`
	IsTest      bool   `json:"is_test"`
	IsConfig    bool   `json:"is_config"`
	IsGenerated bool   `json:"is_generated"`
}

// AnalysisOutput represents analysis results in the output.
type AnalysisOutput struct {
	ChangeType    string   `json:"change_type"`
	RiskLevel     string   `json:"risk_level"`
	AffectedAreas []string `json:"affected_areas"`
	ReviewAxes    []string `json:"review_axes"`
	RelatedFiles  []string `json:"related_files"`
	Warnings      []string `json:"warnings"`
	Summary       string   `json:"summary"`
}
