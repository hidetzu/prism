package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/hidetzu/prism/internal/domain"
	"github.com/hidetzu/prism/internal/fileutil"
)

const protocolVersion = "1"

// pluginOutput represents the JSON schema that provider plugins must produce.
type pluginOutput struct {
	Version      string              `json:"version"`
	Provider     string              `json:"provider"`
	Repository   string              `json:"repository"`
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Author       string              `json:"author"`
	SourceBranch string              `json:"source_branch"`
	TargetBranch string              `json:"target_branch"`
	Description  string              `json:"description"`
	ChangedFiles []pluginChangedFile `json:"changed_files"`
}

type pluginChangedFile struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

// Provider implements provider.Provider by executing an external plugin binary.
type Provider struct {
	name       string
	binaryPath string
	prURL      string
}

// NewProvider creates a plugin-based provider.
func NewProvider(name, binaryPath, prURL string) *Provider {
	return &Provider{
		name:       name,
		binaryPath: binaryPath,
		prURL:      prURL,
	}
}

// Parse returns a minimal PRRef. For plugin providers, the binary does the real parsing.
func (p *Provider) Parse(input string) (domain.PRRef, error) {
	return domain.PRRef{Provider: p.name}, nil
}

// FetchPullRequest executes the plugin binary and converts its JSON output to a domain.PullRequest.
func (p *Provider) FetchPullRequest(ctx context.Context, ref domain.PRRef) (domain.PullRequest, error) {
	cmd := exec.CommandContext(ctx, p.binaryPath, "fetch", p.prURL)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return domain.PullRequest{}, fmt.Errorf(
			"%w: plugin %s failed: %v: %s", domain.ErrProvider, p.name, err, stderr.String(),
		)
	}

	var out pluginOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return domain.PullRequest{}, fmt.Errorf(
			"%w: plugin %s returned invalid JSON: %v", domain.ErrProvider, p.name, err,
		)
	}

	if out.Version != protocolVersion {
		return domain.PullRequest{}, fmt.Errorf(
			"%w: plugin %s uses protocol version %q, expected %q",
			domain.ErrProvider, p.name, out.Version, protocolVersion,
		)
	}

	return convertOutput(out), nil
}

func convertOutput(out pluginOutput) domain.PullRequest {
	files := make([]domain.ChangedFile, len(out.ChangedFiles))
	for i, f := range out.ChangedFiles {
		files[i] = domain.ChangedFile{
			Path:        f.Path,
			Status:      mapFileStatus(f.Status),
			Additions:   f.Additions,
			Deletions:   f.Deletions,
			Language:    fileutil.DetectLanguage(f.Path),
			IsTest:      fileutil.IsTestFile(f.Path),
			IsConfig:    fileutil.IsConfigFile(f.Path),
			IsGenerated: fileutil.IsGeneratedFile(f.Path),
			Patch:       f.Patch,
		}
	}

	return domain.PullRequest{
		Repository:   out.Repository,
		ID:           out.ID,
		Title:        out.Title,
		Author:       out.Author,
		SourceBranch: out.SourceBranch,
		TargetBranch: out.TargetBranch,
		Description:  out.Description,
		ChangedFiles: files,
	}
}

func mapFileStatus(status string) domain.FileStatus {
	switch status {
	case "added":
		return domain.FileStatusAdded
	case "modified":
		return domain.FileStatusModified
	case "removed":
		return domain.FileStatusRemoved
	case "renamed":
		return domain.FileStatusRenamed
	default:
		return domain.FileStatusModified
	}
}
