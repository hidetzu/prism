# prism

[![CI](https://github.com/hidetzu/prism/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/prism/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidetzu/prism)](https://goreportcard.com/report/github.com/hidetzu/prism)
[![Go Reference](https://pkg.go.dev/badge/github.com/hidetzu/prism.svg)](https://pkg.go.dev/github.com/hidetzu/prism)
[![Release](https://img.shields.io/github/v/release/hidetzu/prism)](https://github.com/hidetzu/prism/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Decompose pull requests into stable, reviewable system context

A prism decomposes light into its spectrum. **prism** decomposes pull requests — extracting structure, intent, and risk from raw diffs so that AI reviews become designable and reproducible.

**PR → structured context → stable AI review.**

---

## The problem

AI code reviews are inconsistent — not because models are bad, but because the input is unstructured:

- Raw diffs lack context about *what* changed and *why*
- Reviewers must decide review depth and focus points every time
- Different prompts produce wildly different review quality
- No standardized way to tell AI *what to look for*

prism solves this by compiling PRs into structured, reviewable system context.

---

## What prism does

prism is a **Review Context Compiler** — it decomposes Pull Requests into AI-review-ready system context by:

- Extracting PR metadata and diffs
- Classifying change types (feature, bugfix, refactor, ...)
- Estimating risk level
- Suggesting review axes (security, backward compatibility, error handling, ...)
- Detecting related files (test/source pairs, peer files)
- Generating structured output (JSON / Markdown / text)
- Producing mode-specific review prompts (light / detailed / cross)

---

## The simplest way to use it

```bash
prism analyze <PR_URL> --format json | claude -p "Review this pull request"
```

---

## Quick Start

Requires [Go](https://go.dev/dl/) 1.26 or later.

```bash
go install github.com/hidetzu/prism/cmd/prism@latest
export GITHUB_TOKEN=your-token
prism analyze https://github.com/owner/repo/pull/123 --format json
```

---

## Use Cases

### Analyze a PR

```bash
prism analyze <PR_URL> --format json       # Structured JSON for AI tools
prism analyze <PR_URL> --format markdown   # Human-readable summary
prism analyze <PR_URL> --format text       # Plain text summary
```

### Generate review prompts

```bash
prism prompt <PR_URL> --mode light         # Quick screening
prism prompt <PR_URL> --mode detailed      # Deep review with patches
prism prompt <PR_URL> --mode cross         # Cross-file consistency review
prism prompt <PR_URL> --mode light --lang ja  # Japanese prompt
```

### Pipe to Claude

```bash
prism analyze <PR_URL> --format json | claude -p "Review this pull request"
prism prompt <PR_URL> --mode detailed | claude -p
```

### Debug PR data

```bash
prism fetch <PR_URL> --format json         # Raw PR data (no analysis)
prism fetch <PR_URL> --format text         # Raw PR data as text
```

---

## What you get

<details>
<summary>JSON output example</summary>

```json
{
  "provider": "github",
  "pull_request": {
    "repository": "owner/repo",
    "id": "123",
    "title": "Add retry handling for payment API",
    "author": "example",
    "source_branch": "feature/payment-retry",
    "target_branch": "main",
    "description": "Adds retry logic for transient payment API failures"
  },
  "changed_files": [
    {
      "path": "internal/payment/client.go",
      "status": "modified",
      "additions": 45,
      "deletions": 3,
      "language": "Go",
      "is_test": false,
      "is_config": false,
      "is_generated": false
    }
  ],
  "analysis": {
    "change_type": "feature",
    "risk_level": "medium",
    "affected_areas": ["payment"],
    "review_axes": [
      "error handling",
      "test coverage",
      "edge cases"
    ],
    "related_files": [
      "internal/payment/client_test.go",
      "internal/payment/handler.go",
      "internal/payment/service.go"
    ],
    "warnings": [
      "No test files included in this change"
    ],
    "summary": "feature: Add retry handling for payment API (1 file changed, +45/-3)"
  }
}
```

</details>

<details>
<summary>Markdown output example</summary>

```markdown
# Add retry handling for payment API

## Pull Request

| Field | Value |
|-------|-------|
| Repository | owner/repo |
| PR | #123 |
| Author | example |
| Branch | feature/payment-retry -> main |
| Provider | github |

## Analysis

- **Change Type:** feature
- **Risk Level:** medium
- **Summary:** feature: Add retry handling for payment API (1 file changed, +45/-3)

### Review Axes

- error handling
- test coverage
- edge cases

### Warnings

- No test files included in this change

## Changed Files

| File | Status | +/- | Language |
|------|--------|-----|----------|
| internal/payment/client.go | modified | +45/-3 | Go |
```

</details>

<details>
<summary>Commands reference</summary>

```bash
# Analyze
prism analyze <PR_URL>                     # JSON (default)
prism analyze <PR_URL> --format json       # Structured JSON
prism analyze <PR_URL> --format markdown   # Markdown summary
prism analyze <PR_URL> --format text       # Plain text summary

# Prompt generation
prism prompt <PR_URL> --mode light         # Quick screening prompt
prism prompt <PR_URL> --mode detailed      # Deep review prompt
prism prompt <PR_URL> --mode cross         # Cross-file review prompt
prism prompt <PR_URL> --lang ja            # Japanese prompt
prism prompt <PR_URL> --template my.tmpl   # Custom template

# Debug
prism fetch <PR_URL> --format json         # Raw PR data
prism fetch <PR_URL> --format text         # Raw PR data as text

# Provider selection (available on all commands)
prism analyze <PR_URL> --provider github   # Explicit provider
```

</details>

---

## Prompt Modes

| Mode | Purpose | Depth |
|------|---------|-------|
| `light` | Quick screening for obvious bugs, security issues | Minimal context |
| `detailed` | Implementation review, coverage gaps, edge cases | Full context with patches |
| `cross` | Cross-file consistency, interface contracts | Module structure focus |

---

## Providers

### Supported providers

| Provider | Type | Status |
|----------|------|--------|
| GitHub | Built-in | Supported |
| AWS CodeCommit | Plugin | Supported ([prism-provider-codecommit](https://github.com/hidetzu/prism-provider-codecommit)) |

### Provider selection

By default, prism auto-detects the provider from the PR URL. Use `--provider` to specify explicitly:

```bash
prism analyze https://github.com/owner/repo/pull/123              # auto-detected as GitHub
prism analyze <PR_URL> --provider github                          # explicit GitHub
prism analyze <PR_URL> --provider codecommit                      # explicit CodeCommit (requires plugin)
```

### Plugin providers

External providers are distributed as separate binaries named `prism-provider-<name>` and discovered on PATH. Plugins receive a PR URL and return structured JSON to stdout.

```bash
# Plugin invocation (called by prism internally):
prism-provider-codecommit fetch <PR_URL>
```

See [ADR-0001](docs/adr/0001-provider-plugin-architecture.md) for design details.

---

## Configuration

### GitHub token

Set `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

### Config file

prism loads configuration from `~/.config/prism/config.yaml` (override with `--config` or `PRISM_CONFIG` env var).

```yaml
github_token: ghp_xxxxxxxxxxxx
default_format: json        # json, markdown, text
default_mode: light         # light, detailed, cross
default_lang: en            # en, ja
```

Environment variables override config file values. `GITHUB_TOKEN` always takes precedence.

### Custom templates

Use `--template` to provide a custom [Go template](https://pkg.go.dev/text/template) for prompt output:

```bash
prism prompt <PR_URL> --template review.tmpl
```

Available template variables:

| Variable | Type | Description |
|----------|------|-------------|
| `.PR` | PullRequest | PR metadata and changed files |
| `.Analysis` | AnalysisResult | Classification, risk, review axes |
| `.Mode` | string | Prompt mode (light/detailed/cross) |
| `.Lang` | string | Language code (en/ja) |
| `.SystemPrompt` | string | Built-in system prompt for the mode |

Example template:

```
Review: {{.PR.Title}} ({{.Analysis.ChangeType}}, risk: {{.Analysis.RiskLevel}})

{{range .Analysis.ReviewAxes}}- Focus: {{.}}
{{end}}
{{range .PR.ChangedFiles}}- {{.Path}} ({{.Status}}, +{{.Additions}}/-{{.Deletions}})
{{end}}
```

### Language support

Use `--lang` to switch prompt language:

```bash
prism prompt <PR_URL> --lang ja    # Japanese system prompts
```

Supported: `en` (English, default), `ja` (Japanese)

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments (bad URL, invalid flag values) |
| 3 | Provider error (GitHub API failure, auth error) |
| 4 | Analysis error |

---

## Development

```bash
make build    # Build binary to bin/prism
make test     # Run all tests with -v -race
make lint     # golangci-lint
make vet      # go vet
make clean    # Remove bin/
```

## Roadmap

- **v0.1.0** — GitHub provider, analyze/prompt/fetch commands, JSON/Markdown/text output, light/detailed/cross modes, config/lang/template support, exit codes
- **v0.2.0** — Provider plugin architecture, `--provider` flag, AWS CodeCommit provider
- **v0.3.0** — Policy files, custom review axes, project-specific rules
- **v0.4.0+** — Review policy as code, SARIF output, metrics, IDE/CI integration

## License

MIT
