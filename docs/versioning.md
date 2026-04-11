# Versioning Policy

prism follows [Semantic Versioning 2.0.0](https://semver.org/).

Two surfaces have stability guarantees:

1. **CLI JSON output** (`prism analyze --format json`) — a contract with shell scripts, AI tools, CI pipelines
2. **`pkg/prism` public API** — a contract with Go library consumers (prism-api, editor plugins, CI tools)

Both follow the same semantic versioning rules described below.

## JSON Output Stability

The JSON output from `prism analyze` is a contract with downstream consumers (AI tools, CI pipelines, scripts).

### Within a major version (e.g., v1.x.x)

**Allowed:**
- Adding new fields to objects
- Adding new enum values (e.g., new change types, new review axes)
- Adding new commands or flags

**Not allowed:**
- Renaming existing fields
- Removing existing fields
- Changing the type of existing fields
- Changing the nesting structure of existing fields

### Major version bumps

Breaking changes to JSON output require a major version bump. When this happens:

- Document all breaking changes in release notes
- Provide a migration guide if the changes are complex

## Public Library API Stability (`pkg/prism`)

The `pkg/prism` package is a contract with Go library consumers.

### Within a major version

**Allowed:**
- Adding new fields to `AnalyzeOptions` (as optional)
- Adding new fields to `Result`, `PRInfo`, `AnalysisResult`, `ChangedFile`
- Adding new sentinel errors
- Adding new exported functions

**Not allowed:**
- Changing function signatures of `Analyze` or `Prompt`
- Renaming or removing existing fields
- Changing the type of existing fields
- Changing default behavior in backwards-incompatible ways

See [ADR-0002](adr/0002-public-api-boundary.md) for the full compatibility policy and the difference between `pkg/prism.Result` and CLI JSON output.

## Pre-1.0

During v0.x development, both the CLI JSON schema and the `pkg/prism` public API may change between minor versions. Breaking changes will be documented in release notes.

## Golden Tests

JSON output is verified by golden tests in `testdata/`. If golden test files need updating due to output changes, the change must be reviewed explicitly to ensure backward compatibility is intentional.
