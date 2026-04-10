package domain

// ChangeType represents the classified type of a pull request change.
type ChangeType string

const (
	ChangeTypeFeature          ChangeType = "feature"
	ChangeTypeBugfix           ChangeType = "bugfix"
	ChangeTypeRefactor         ChangeType = "refactor"
	ChangeTypeTestOnly         ChangeType = "test-only"
	ChangeTypeDocsOnly         ChangeType = "docs-only"
	ChangeTypeConfigChange     ChangeType = "config-change"
	ChangeTypeDependencyUpdate ChangeType = "dependency-update"
	ChangeTypeInfraChange      ChangeType = "infra-change"
)

// RiskLevel represents the estimated risk of a change.
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// ReviewAxis represents a suggested review focus point.
type ReviewAxis string

const (
	ReviewAxisErrorHandling       ReviewAxis = "error handling"
	ReviewAxisBackwardCompat      ReviewAxis = "backward compatibility"
	ReviewAxisTestCoverage        ReviewAxis = "test coverage"
	ReviewAxisPerformance         ReviewAxis = "performance"
	ReviewAxisSecurity            ReviewAxis = "security"
	ReviewAxisConfigSafety        ReviewAxis = "configuration safety"
	ReviewAxisEdgeCases           ReviewAxis = "edge cases"
	ReviewAxisReadability         ReviewAxis = "readability"
	ReviewAxisSeparationOfConcern ReviewAxis = "separation of concerns"
)

// FileStatus represents the change status of a file.
type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusRemoved  FileStatus = "removed"
	FileStatusRenamed  FileStatus = "renamed"
)

// PromptMode represents the review prompt generation mode.
type PromptMode string

const (
	PromptModeLight    PromptMode = "light"
	PromptModeDetailed PromptMode = "detailed"
	PromptModeCross    PromptMode = "cross"
)

// PRRef is a provider-agnostic reference to a pull request.
type PRRef struct {
	Provider string
	Owner    string
	Repo     string
	Number   int
}

// PullRequest represents the normalized metadata and content of a pull request.
type PullRequest struct {
	Repository   string
	ID           string
	Title        string
	Author       string
	SourceBranch string
	TargetBranch string
	Description  string
	ChangedFiles []ChangedFile
}

// ChangedFile represents a single file changed in a pull request.
type ChangedFile struct {
	Path        string
	Status      FileStatus
	Additions   int
	Deletions   int
	Language    string
	IsTest      bool
	IsConfig    bool
	IsGenerated bool
	Patch       string
}

// AnalysisResult holds the output of the classification and analysis pipeline.
type AnalysisResult struct {
	ChangeType    ChangeType
	RiskLevel     RiskLevel
	AffectedAreas []string
	ReviewAxes    []ReviewAxis
	RelatedFiles  []string
	Warnings      []string
	Summary       string
}

// PromptBundle holds the assembled prompt for AI review consumption.
type PromptBundle struct {
	Mode         PromptMode
	SystemPrompt string
	UserPrompt   string
}
