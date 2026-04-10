package prompt

import (
	"bytes"
	"os"
	"text/template"

	"github.com/hidetzu/prism/internal/domain"
)

// TemplateData is the data passed to custom templates.
type TemplateData struct {
	Mode         string
	Lang         string
	PR           domain.PullRequest
	Analysis     domain.AnalysisResult
	SystemPrompt string
}

// RenderFromTemplate renders a PromptBundle using a custom template file.
func RenderFromTemplate(path string, mode domain.PromptMode, pr domain.PullRequest, result domain.AnalysisResult, lang string) (domain.PromptBundle, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.PromptBundle{}, err
	}

	tmpl, err := template.New("custom").Parse(string(content))
	if err != nil {
		return domain.PromptBundle{}, err
	}

	sp := systemPrompts(lang)
	var sysPrompt string
	switch mode {
	case domain.PromptModeDetailed:
		sysPrompt = sp.detailed
	case domain.PromptModeCross:
		sysPrompt = sp.cross
	default:
		sysPrompt = sp.light
	}

	data := TemplateData{
		Mode:         string(mode),
		Lang:         lang,
		PR:           pr,
		Analysis:     result,
		SystemPrompt: sysPrompt,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return domain.PromptBundle{}, err
	}

	return domain.PromptBundle{
		Mode:         mode,
		SystemPrompt: sysPrompt,
		UserPrompt:   buf.String(),
	}, nil
}
