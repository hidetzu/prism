# JSON Output Reference

This document describes the JSON output of the `prism analyze` CLI command.

The CLI calls [`pkg/prism.Analyze()`](../pkg/prism) internally and serializes the returned `Result` as JSON. The CLI JSON output and `pkg/prism.Result` are guaranteed to be byte-identical (verified by golden tests in `internal/formatter/testdata/`).

## `prism analyze` output

### Top-level structure

| Field | Type | Description |
|-------|------|-------------|
| `pull_request` | object | PR metadata (provider, repository, identifiers, branches, URL) |
| `analysis` | object | Analysis results (change type, risk, review axes) |
| `changed_files` | array | List of changed files |

### `pull_request`

| Field | Type | Description |
|-------|------|-------------|
| `provider` | string | PR source (`"github"`, `"codecommit"`, etc.) |
| `repository` | string | Repository identifier (e.g., `"owner/repo"`) |
| `id` | string | PR number/ID |
| `title` | string | PR title |
| `author` | string | PR author |
| `source_branch` | string | Source branch name (omitted if empty) |
| `target_branch` | string | Target branch name (omitted if empty) |
| `url` | string | Canonical PR URL |

> **Note:** PR description is not included in the JSON output. Description text can be large and is rarely used by programmatic consumers. Use `prism fetch` if you need the raw description.

### `changed_files[]`

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | File path |
| `status` | string | Change status (`"added"`, `"modified"`, `"removed"`, `"renamed"`) |
| `additions` | int | Lines added |
| `deletions` | int | Lines deleted |
| `language` | string | Detected language (omitted if unknown) |
| `is_test` | bool | Test file flag (omitted if false) |
| `is_config` | bool | Config file flag (omitted if false) |
| `is_generated` | bool | Generated file flag (omitted if false) |
| `patch` | string | Unified diff content (omitted; populated only when explicitly requested via library API) |

### `analysis`

| Field | Type | Description |
|-------|------|-------------|
| `change_type` | string | Classified change type (see below) |
| `risk_level` | string | `"low"`, `"medium"`, or `"high"` |
| `affected_areas` | string[] | Affected domain areas (omitted if empty) |
| `review_axes` | string[] | Suggested review focus points (omitted if empty) |
| `related_files` | string[] | Files related to the change (omitted if empty) |
| `warnings` | string[] | Notable concerns (omitted if empty) |
| `summary` | string | Brief description of the change (omitted if empty) |

### Change types

| Value | Description |
|-------|-------------|
| `feature` | New functionality |
| `bugfix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `test-only` | Only test files changed |
| `docs-only` | Only documentation changed |
| `config-change` | Configuration file changes |
| `dependency-update` | Dependency version changes |
| `infra-change` | Infrastructure/CI changes |

### Review axes

Possible values: `error handling`, `backward compatibility`, `test coverage`, `performance`, `security`, `configuration safety`, `edge cases`, `readability`, `separation of concerns`

---

## Versioning

JSON output follows [Semantic Versioning](versioning.md):

- **Patch/Minor**: New fields may be added (backward compatible)
- **Major**: Fields may be renamed, removed, or change type
- Consumers should ignore unknown fields for forward compatibility

## Breaking changes in v0.3.0

The JSON output structure changed in v0.3.0 to align with `pkg/prism.Result`:

- `provider` moved from top-level to `pull_request.provider`
- `pull_request.description` removed (use `prism fetch` for raw description)
- `pull_request.url` added (canonical PR URL)
- `changed_files[].patch` no longer included by default
- Empty/zero fields are now omitted from output (`omitempty` semantics)

If you have downstream tooling parsing the v0.2.x output, update it to expect the new structure.
