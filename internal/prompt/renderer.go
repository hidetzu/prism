package prompt

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hidetzu/prism/internal/domain"
)

// Render generates a PromptBundle for the given mode and language.
// Supported languages: "en" (default), "ja".
func Render(mode domain.PromptMode, pr domain.PullRequest, result domain.AnalysisResult, lang string) domain.PromptBundle {
	sp := systemPrompts(lang)
	switch mode {
	case domain.PromptModeDetailed:
		return renderDetailed(pr, result, sp.detailed)
	case domain.PromptModeCross:
		return renderCross(pr, result, sp.cross)
	default:
		return renderLight(pr, result, sp.light)
	}
}

func renderLight(pr domain.PullRequest, result domain.AnalysisResult, sysPrompt string) domain.PromptBundle {
	var user strings.Builder

	fmt.Fprintf(&user, "## Pull Request: %s\n\n", pr.Title)
	fmt.Fprintf(&user, "Repository: %s | PR #%s | Author: %s\n", pr.Repository, pr.ID, pr.Author)
	fmt.Fprintf(&user, "Branch: %s -> %s\n\n", pr.SourceBranch, pr.TargetBranch)

	fmt.Fprintf(&user, "**Change Type:** %s | **Risk:** %s\n\n", result.ChangeType, result.RiskLevel)

	if result.Summary != "" {
		fmt.Fprintf(&user, "**Summary:** %s\n\n", result.Summary)
	}

	if len(result.ReviewAxes) > 0 {
		user.WriteString("**Focus on:**\n")
		for _, axis := range result.ReviewAxes {
			fmt.Fprintf(&user, "- %s\n", axis)
		}
		user.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		user.WriteString("**Warnings:**\n")
		for _, w := range result.Warnings {
			fmt.Fprintf(&user, "- %s\n", w)
		}
		user.WriteString("\n")
	}

	// File summary (names only, no patches).
	user.WriteString("**Changed Files:**\n")
	for _, f := range pr.ChangedFiles {
		fmt.Fprintf(&user, "- %s (%s, +%d/-%d)\n", f.Path, f.Status, f.Additions, f.Deletions)
	}

	return domain.PromptBundle{
		Mode:         domain.PromptModeLight,
		SystemPrompt: sysPrompt,
		UserPrompt:   user.String(),
	}
}

func renderDetailed(pr domain.PullRequest, result domain.AnalysisResult, sysPrompt string) domain.PromptBundle {
	var user strings.Builder

	fmt.Fprintf(&user, "## Pull Request: %s\n\n", pr.Title)
	fmt.Fprintf(&user, "Repository: %s | PR #%s | Author: %s\n", pr.Repository, pr.ID, pr.Author)
	fmt.Fprintf(&user, "Branch: %s -> %s\n\n", pr.SourceBranch, pr.TargetBranch)

	if pr.Description != "" {
		fmt.Fprintf(&user, "### Description\n\n%s\n\n", pr.Description)
	}

	fmt.Fprintf(&user, "**Change Type:** %s | **Risk:** %s\n\n", result.ChangeType, result.RiskLevel)

	if result.Summary != "" {
		fmt.Fprintf(&user, "**Summary:** %s\n\n", result.Summary)
	}

	if len(result.AffectedAreas) > 0 {
		fmt.Fprintf(&user, "**Affected Areas:** %s\n\n", strings.Join(result.AffectedAreas, ", "))
	}

	if len(result.ReviewAxes) > 0 {
		user.WriteString("**Review Axes:**\n")
		for _, axis := range result.ReviewAxes {
			fmt.Fprintf(&user, "- %s\n", axis)
		}
		user.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		user.WriteString("**Warnings:**\n")
		for _, w := range result.Warnings {
			fmt.Fprintf(&user, "- %s\n", w)
		}
		user.WriteString("\n")
	}

	// Full file details with patches.
	user.WriteString("### Changed Files\n\n")
	for _, f := range pr.ChangedFiles {
		flags := formatFlags(f)
		fmt.Fprintf(&user, "#### %s%s\n\n", f.Path, flags)
		fmt.Fprintf(&user, "Status: %s | Language: %s | +%d/-%d\n\n", f.Status, f.Language, f.Additions, f.Deletions)
		if f.Patch != "" {
			fmt.Fprintf(&user, "```diff\n%s\n```\n\n", f.Patch)
		}
	}

	return domain.PromptBundle{
		Mode:         domain.PromptModeDetailed,
		SystemPrompt: sysPrompt,
		UserPrompt:   user.String(),
	}
}

func formatFlags(f domain.ChangedFile) string {
	var flags []string
	if f.IsTest {
		flags = append(flags, "test")
	}
	if f.IsConfig {
		flags = append(flags, "config")
	}
	if f.IsGenerated {
		flags = append(flags, "generated")
	}
	if len(flags) == 0 {
		return ""
	}
	return " (" + strings.Join(flags, ", ") + ")"
}

type sysPromptSet struct {
	light    string
	detailed string
	cross    string
}

func systemPrompts(lang string) sysPromptSet {
	if lang == "ja" {
		return sysPromptSet{
			light:    lightSystemPromptJA,
			detailed: detailedSystemPromptJA,
			cross:    crossSystemPromptJA,
		}
	}
	return sysPromptSet{
		light:    lightSystemPrompt,
		detailed: detailedSystemPrompt,
		cross:    crossSystemPrompt,
	}
}

const lightSystemPrompt = `You are a code reviewer performing a quick screening of a pull request.
Focus on identifying obvious issues, risks, and areas that need attention.
Be concise. Prioritize the review axes provided.
If everything looks reasonable, say so briefly.`

const detailedSystemPrompt = `You are a code reviewer performing a thorough review of a pull request.
Examine the changes carefully, paying attention to:
- Correctness and potential bugs
- The review axes highlighted in the context
- Error handling and edge cases
- Code quality and maintainability

Provide specific, actionable feedback referencing file names and line context.
Organize your review by file or by concern, whichever is clearer.`

const crossSystemPrompt = `You are a code reviewer performing a cross-file consistency review of a pull request.
Focus on:
- Interface contracts: do callers and implementations agree on types, error handling, and behavior?
- Configuration consistency: are config keys, environment variables, and feature flags used consistently?
- Test coverage alignment: do test files cover the changed behavior? Are integration points tested?
- Module boundaries: do the changes respect existing package boundaries and dependency directions?

Identify inconsistencies across files rather than line-level bugs.
Reference specific file pairs when flagging issues.`

func renderCross(pr domain.PullRequest, result domain.AnalysisResult, sysPrompt string) domain.PromptBundle {
	var user strings.Builder

	fmt.Fprintf(&user, "## Pull Request: %s\n\n", pr.Title)
	fmt.Fprintf(&user, "Repository: %s | PR #%s | Author: %s\n", pr.Repository, pr.ID, pr.Author)
	fmt.Fprintf(&user, "Branch: %s -> %s\n\n", pr.SourceBranch, pr.TargetBranch)

	fmt.Fprintf(&user, "**Change Type:** %s | **Risk:** %s\n\n", result.ChangeType, result.RiskLevel)

	if len(result.AffectedAreas) > 0 {
		fmt.Fprintf(&user, "**Affected Areas:** %s\n\n", strings.Join(result.AffectedAreas, ", "))
	}

	// Module structure: group files by directory.
	modules := groupByDirectory(pr.ChangedFiles)
	user.WriteString("### Module Structure\n\n")
	for _, m := range modules {
		fmt.Fprintf(&user, "**%s/**\n", m.dir)
		for _, f := range m.files {
			role := fileRole(f)
			fmt.Fprintf(&user, "  - %s [%s] (%s, +%d/-%d)\n", f.Path, role, f.Status, f.Additions, f.Deletions)
		}
		user.WriteString("\n")
	}

	// Related files (not changed but potentially affected).
	if len(result.RelatedFiles) > 0 {
		user.WriteString("### Related Files (not changed)\n\n")
		for _, rf := range result.RelatedFiles {
			fmt.Fprintf(&user, "- %s\n", rf)
		}
		user.WriteString("\n")
	}

	// Cross-file relationships.
	user.WriteString("### Cross-File Relationships\n\n")

	// Test ↔ source pairs.
	testPairs := findTestSourcePairs(pr.ChangedFiles)
	if len(testPairs) > 0 {
		user.WriteString("**Test ↔ Source pairs in this change:**\n")
		for _, pair := range testPairs {
			fmt.Fprintf(&user, "  - %s ↔ %s\n", pair[0], pair[1])
		}
		user.WriteString("\n")
	}

	// Config files.
	var configFiles []string
	for _, f := range pr.ChangedFiles {
		if f.IsConfig {
			configFiles = append(configFiles, f.Path)
		}
	}
	if len(configFiles) > 0 {
		user.WriteString("**Configuration files changed:**\n")
		for _, cf := range configFiles {
			fmt.Fprintf(&user, "  - %s\n", cf)
		}
		user.WriteString("\n")
	}

	if len(result.ReviewAxes) > 0 {
		user.WriteString("**Review Axes:**\n")
		for _, axis := range result.ReviewAxes {
			fmt.Fprintf(&user, "- %s\n", axis)
		}
		user.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		user.WriteString("**Warnings:**\n")
		for _, w := range result.Warnings {
			fmt.Fprintf(&user, "- %s\n", w)
		}
		user.WriteString("\n")
	}

	return domain.PromptBundle{
		Mode:         domain.PromptModeCross,
		SystemPrompt: sysPrompt,
		UserPrompt:   user.String(),
	}
}

type moduleGroup struct {
	dir   string
	files []domain.ChangedFile
}

func groupByDirectory(files []domain.ChangedFile) []moduleGroup {
	dirMap := make(map[string][]domain.ChangedFile)
	var dirs []string
	for _, f := range files {
		dir := filepath.Dir(f.Path)
		if _, exists := dirMap[dir]; !exists {
			dirs = append(dirs, dir)
		}
		dirMap[dir] = append(dirMap[dir], f)
	}
	sort.Strings(dirs)

	groups := make([]moduleGroup, 0, len(dirs))
	for _, d := range dirs {
		groups = append(groups, moduleGroup{dir: d, files: dirMap[d]})
	}
	return groups
}

func fileRole(f domain.ChangedFile) string {
	switch {
	case f.IsTest:
		return "test"
	case f.IsConfig:
		return "config"
	case f.IsGenerated:
		return "generated"
	default:
		return "source"
	}
}

func findTestSourcePairs(files []domain.ChangedFile) [][2]string {
	sources := make(map[string]bool)
	tests := make(map[string]bool)

	for _, f := range files {
		if f.IsTest {
			tests[f.Path] = true
		} else {
			sources[f.Path] = true
		}
	}

	var pairs [][2]string
	for src := range sources {
		ext := filepath.Ext(src)
		if ext == ".go" {
			testFile := strings.TrimSuffix(src, ext) + "_test" + ext
			if tests[testFile] {
				pairs = append(pairs, [2]string{src, testFile})
			}
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })
	return pairs
}

const lightSystemPromptJA = `あなたはプルリクエストの簡易スクリーニングを行うコードレビュアーです。
明らかな問題、リスク、注意が必要な箇所を特定することに集中してください。
簡潔に回答してください。提供されたレビュー軸を優先してください。
問題がなさそうであれば、その旨を簡潔に述べてください。`

const detailedSystemPromptJA = `あなたはプルリクエストの詳細なレビューを行うコードレビュアーです。
以下の点に注意して変更を丁寧に確認してください：
- 正確性と潜在的なバグ
- コンテキストで示されたレビュー軸
- エラーハンドリングとエッジケース
- コード品質と保守性

ファイル名と行のコンテキストを参照して、具体的で実行可能なフィードバックを提供してください。
ファイル別または懸念事項別に整理してください。`

const crossSystemPromptJA = `あなたはプルリクエストのクロスファイル整合性レビューを行うコードレビュアーです。
以下の点に集中してください：
- インターフェース契約：呼び出し元と実装は型、エラーハンドリング、振る舞いで合意しているか？
- 設定の一貫性：設定キー、環境変数、フィーチャーフラグは一貫して使われているか？
- テストカバレッジの整合性：テストファイルは変更された振る舞いをカバーしているか？
- モジュール境界：変更は既存のパッケージ境界と依存方向を尊重しているか？

行レベルのバグではなく、ファイル間の不整合を特定してください。
問題を指摘する際は、具体的なファイルペアを参照してください。`
