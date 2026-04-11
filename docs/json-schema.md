# JSON Output Reference

This document describes the JSON output of the `prism analyze` CLI command.

> **Note:** Library consumers using [`pkg/prism`](../pkg/prism) receive a `Result` struct that is structurally similar but not identical to this CLI JSON output. See [ADR-0002](adr/0002-public-api-boundary.md) for the differences. The CLI JSON schema and `pkg/prism.Result` will be unified in Phase 2 when the CLI is refactored to call `pkg/prism` internally.

## `prism analyze` output

### Top-level structure

| Field | Type | Description |
|-------|------|-------------|
| `provider` | string | PR source (`"github"`, `"codecommit"`) |
| `pull_request` | object | PR metadata |
| `changed_files` | array | List of changed files |
| `analysis` | object | Analysis results |

### `pull_request`

| Field | Type | Description |
|-------|------|-------------|
| `repository` | string | Repository identifier (e.g., `"owner/repo"`) |
| `id` | string | PR number/ID |
| `title` | string | PR title |
| `author` | string | PR author |
| `source_branch` | string | Source branch name |
| `target_branch` | string | Target branch name |
| `description` | string | PR description body |

### `changed_files[]`

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | File path |
| `status` | string | Change status (`"added"`, `"modified"`, `"removed"`, `"renamed"`) |
| `additions` | int | Lines added |
| `deletions` | int | Lines deleted |
| `language` | string | Detected language |
| `is_test` | bool | Whether this is a test file |
| `is_config` | bool | Whether this is a config file |
| `is_generated` | bool | Whether this is a generated file |

### `analysis`

| Field | Type | Description |
|-------|------|-------------|
| `change_type` | string | Classified change type (see below) |
| `risk_level` | string | `"low"`, `"medium"`, or `"high"` |
| `affected_areas` | string[] | Affected domain areas |
| `review_axes` | string[] | Suggested review focus points |
| `related_files` | string[] | Files related to the change |
| `warnings` | string[] | Notable concerns |
| `summary` | string | Brief description of the change |

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
