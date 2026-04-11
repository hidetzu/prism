# CLI Specification

## Global Flags

The following flag is available on all subcommands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--provider` | string | (auto-detect) | Provider name (e.g. `github`, `codecommit`); auto-detected from URL if omitted |

When `--provider` is specified, URL-based auto-detection is skipped and the given provider is used directly.

---

## Commands

### `prism analyze`

Fetches a PR, runs analysis, and outputs structured results.

```bash
prism analyze <PR_URL> [flags]
```

#### Input

- Positional: PR URL (e.g., `https://github.com/owner/repo/pull/123`)

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | `json\|markdown\|text` | `json` | Output format |
| `--config` | string | `~/.config/prism/config.yaml` | Config file path |

#### Output

Returns an `AnalysisResult` in the specified format. See [JSON Schema](json-schema.md) for field definitions.

---

### `prism prompt`

Generates a review prompt for AI consumption.

```bash
prism prompt <PR_URL> [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--mode` | `light\|detailed\|cross` | `light` | Prompt mode |
| `--format` | `text\|markdown\|json` | `text` | Output format |
| `--lang` | `en\|ja` | `en` | Prompt language |
| `--template` | string | built-in | Custom Go template path |
| `--config` | string | `~/.config/prism/config.yaml` | Config file path |

#### Prompt Modes

**light** — Quick screening for obvious issues.
- Includes: PR title, summary, high-level diff summary, top review axes
- Use case: first-pass triage

**detailed** — Thorough implementation review.
- Includes: summary, changed files, related files, warnings, expanded review axes, patch excerpts
- Use case: full code review

**cross** — Cross-file consistency analysis.
- Includes: module structure, changed/related files, test/source pairs, config files, interface relationships
- Use case: architectural review, integration review

---

### `prism fetch`

Fetches raw PR data for debugging. No analysis is performed.

```bash
prism fetch <PR_URL> [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | `json\|text` | `json` | Output format |
| `--config` | string | `~/.config/prism/config.yaml` | Config file path |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments (bad URL, invalid flag values) |
| 3 | Provider error (API failure, auth) |
| 4 | Analysis error |

---

## Configuration

### Config file

Loaded from `~/.config/prism/config.yaml` by default. Override with `--config` flag or `PRISM_CONFIG` environment variable.

```yaml
github_token: ghp_xxxxxxxxxxxx
default_format: json
default_mode: light
default_lang: en
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub API authentication token (overrides config file) |
| `PRISM_CONFIG` | Override config file path |

### Custom Templates

The `--template` flag accepts a path to a Go [text/template](https://pkg.go.dev/text/template) file. The template receives a `TemplateData` struct with fields: `PR`, `Analysis`, `Mode`, `Lang`, `SystemPrompt`.
