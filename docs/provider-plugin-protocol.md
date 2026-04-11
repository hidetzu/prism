# Provider Plugin Protocol v1

This document defines the contract between `prism` and provider plugin binaries. Any binary that conforms to this protocol can serve as a prism provider.

## Protocol Version

Current version: **1**

The `version` field in plugin output is required. `prism` will reject output with a mismatched version.

---

## Invocation

`prism` invokes a provider plugin as a subprocess:

```
prism-provider-<name> fetch <PR_URL>
```

### Arguments

| Position | Value | Description |
|----------|-------|-------------|
| 1 | `fetch` | Subcommand (currently the only supported command) |
| 2 | PR URL | The raw PR URL as provided by the user |

### Environment

The plugin inherits the parent process environment. Plugins may use environment variables for authentication (e.g., `AWS_PROFILE`, `GITHUB_TOKEN`).

`prism` does NOT pass authentication credentials to plugins. Each plugin is responsible for its own authentication.

---

## Output

### stdout — Structured JSON (machine-readable)

The plugin MUST write a single JSON object to stdout. `prism` parses only stdout.

```json
{
  "version": "1",
  "provider": "codecommit",
  "repository": "my-repo",
  "id": "42",
  "title": "Add retry handling for payment API",
  "author": "dev",
  "source_branch": "feature/payment-retry",
  "target_branch": "main",
  "description": "Adds retry logic for transient payment API failures",
  "changed_files": [
    {
      "path": "internal/payment/client.go",
      "status": "modified",
      "additions": 45,
      "deletions": 3,
      "patch": "@@ -1,3 +1,45 @@\n+func retry() { ... }"
    },
    {
      "path": "internal/payment/client_test.go",
      "status": "added",
      "additions": 80,
      "deletions": 0,
      "patch": "@@ -0,0 +1,80 @@\n+package payment_test ..."
    }
  ]
}
```

### stderr — Debug/diagnostic output (human-readable)

The plugin MAY write diagnostic messages to stderr. `prism` captures stderr and includes it in error messages when the plugin fails. stderr is NOT parsed as structured data.

---

## Field Reference

### Top-level fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | **yes** | Protocol version. Must be `"1"`. |
| `provider` | string | **yes** | Provider name (e.g., `"github"`, `"codecommit"`). |
| `repository` | string | **yes** | Repository identifier (e.g., `"owner/repo"`, `"my-repo"`). |
| `id` | string | **yes** | Pull request identifier (e.g., `"123"`, `"42"`). |
| `title` | string | **yes** | Pull request title. |
| `author` | string | **yes** | Author name or identifier. |
| `source_branch` | string | **yes** | Source branch name. |
| `target_branch` | string | **yes** | Target (base) branch name. |
| `description` | string | **yes** | Pull request description body. May be empty string `""`. |
| `changed_files` | array | **yes** | List of changed file objects. May be empty array `[]`. |

### changed_files element

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | string | **yes** | File path relative to repository root. |
| `status` | string | **yes** | One of: `"added"`, `"modified"`, `"removed"`, `"renamed"`. |
| `additions` | integer | **yes** | Number of added lines. Must be >= 0. |
| `deletions` | integer | **yes** | Number of deleted lines. Must be >= 0. |
| `patch` | string | **yes** | Unified diff patch content. May be empty string `""` for binary files. |

### Notes on field values

- **`version`**: `prism` validates this field. Unknown versions are rejected with an error.
- **`status`**: Unknown values are treated as `"modified"`. Plugins SHOULD use the defined values.
- **`patch`**: Should be in unified diff format. `prism` passes this through to output without parsing.
- **`description`**: Use `""` (empty string) if no description exists. Do NOT omit the field.
- **`changed_files`**: Use `[]` (empty array) if the PR has no file changes. Do NOT omit the field.

---

## Exit Codes

| Exit Code | Meaning | prism behavior |
|-----------|---------|----------------|
| 0 | Success | Parse stdout JSON |
| non-zero | Failure | Report error with stderr content |

When exit code is non-zero, `prism` wraps the error as a provider error (exit code 3).

---

## Plugin Discovery

`prism` discovers plugins by searching for executables on `PATH`:

```
prism-provider-<name>
```

The `<name>` corresponds to the provider name used with `--provider <name>` or auto-detected from the URL.

### Naming convention

| Provider | Binary name |
|----------|-------------|
| GitHub | `prism-provider-github` |
| CodeCommit | `prism-provider-codecommit` |
| GitLab | `prism-provider-gitlab` |
| Bitbucket | `prism-provider-bitbucket` |

---

## Context and Timeouts

`prism` may terminate the plugin process if a context deadline is exceeded (e.g., user cancellation, timeout). Plugins SHOULD handle `SIGTERM` / `SIGINT` gracefully.

---

## File Enrichment

`prism` automatically enriches each `changed_files` entry after receiving plugin output:

- `language` — detected from file extension
- `is_test` — detected from file name and path patterns
- `is_config` — detected from file name patterns
- `is_generated` — detected from file name patterns

Plugins do NOT need to provide these fields. They are computed by `prism`.

---

## Versioning Policy

- Protocol version `"1"` is the initial stable version.
- Additive changes (new optional fields) do NOT require a version bump.
- Breaking changes (removing fields, changing types, changing semantics) require a new version.
- `prism` will support multiple protocol versions simultaneously when needed.

---

## Example: Minimal Plugin (shell script)

```bash
#!/bin/sh
cat <<'EOF'
{
  "version": "1",
  "provider": "example",
  "repository": "my-org/my-repo",
  "id": "1",
  "title": "Example PR",
  "author": "developer",
  "source_branch": "feature",
  "target_branch": "main",
  "description": "",
  "changed_files": []
}
EOF
```

Save as `prism-provider-example`, make executable, and place on PATH.
