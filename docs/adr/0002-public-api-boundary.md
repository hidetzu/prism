# ADR-0002: Public API Boundary

## Status
Accepted

## Date
2026-04-11

## Context

prism is currently structured as a CLI tool with all logic under `internal/`. Go's `internal` package rule makes these packages inaccessible from outside this module.

Planned future uses of prism include:

- **prism-api** — HTTP service wrapping prism for non-CLI consumers (Fly.io deployment planned)
- **IDE / editor integrations** — direct library usage from Go-based editor extensions
- **GitHub Actions / CI plugins** — custom actions embedding prism logic
- **Third-party tools** — projects building on top of prism's analysis

All of these need to call prism's core functionality as a library. The current `internal/` structure makes this impossible. At the same time, exposing every internal package as public would freeze prism's design and make future refactoring painful.

We need to decide **what to make public**, **what to keep private**, and **what compatibility guarantees to offer**.

---

## Decision

prism exposes a **single high-level public package**: `pkg/prism`.

This package provides two functions that cover the primary use cases (`Analyze` and `Prompt`), along with stable input and output types.

### Public API

```go
package prism

import (
    "context"
    "errors"
)

// AnalyzeOptions configures a prism Analyze or Prompt call.
type AnalyzeOptions struct {
    // Provider is the provider name (e.g. "github", "codecommit").
    // If empty, the provider is auto-detected from PRURL.
    Provider string

    // PRURL is the pull request URL.
    PRURL string

    // GitHubToken is the authentication token for the GitHub provider.
    // Not used for plugin providers that manage their own authentication.
    GitHubToken string

    // IncludePatches controls whether ChangedFile.Patch is populated.
    // Default: false (patches excluded to keep responses lightweight).
    IncludePatches bool

    // Mode is the prompt mode: "light", "detailed", or "cross".
    // Only used by Prompt. Defaults to "light" if empty.
    Mode string

    // Language is the prompt language: "en" or "ja".
    // Only used by Prompt. Defaults to "en" if empty.
    Language string
}

// PRInfo is the minimal pull request metadata returned with every Result.
type PRInfo struct {
    Provider     string `json:"provider"`
    Repository   string `json:"repository"`
    ID           string `json:"id"`
    Title        string `json:"title"`
    Author       string `json:"author"`
    SourceBranch string `json:"source_branch,omitempty"`
    TargetBranch string `json:"target_branch,omitempty"`
    URL          string `json:"url"`
}

// ChangedFile represents a file changed in the pull request.
type ChangedFile struct {
    Path        string `json:"path"`
    Status      string `json:"status"`
    Additions   int    `json:"additions"`
    Deletions   int    `json:"deletions"`
    Language    string `json:"language,omitempty"`
    IsTest      bool   `json:"is_test,omitempty"`
    IsConfig    bool   `json:"is_config,omitempty"`
    IsGenerated bool   `json:"is_generated,omitempty"`
    // Patch is populated only when AnalyzeOptions.IncludePatches is true.
    Patch string `json:"patch,omitempty"`
}

// AnalysisResult holds the structured analysis output.
type AnalysisResult struct {
    ChangeType    string   `json:"change_type"`
    RiskLevel     string   `json:"risk_level"`
    AffectedAreas []string `json:"affected_areas,omitempty"`
    ReviewAxes    []string `json:"review_axes,omitempty"`
    RelatedFiles  []string `json:"related_files,omitempty"`
    Warnings      []string `json:"warnings,omitempty"`
    Summary       string   `json:"summary,omitempty"`
}

// Result is the complete output of Analyze.
// It is structurally similar to the existing CLI JSON schema with documented differences.
type Result struct {
    PR       PRInfo         `json:"pull_request"`
    Analysis AnalysisResult `json:"analysis"`
    Files    []ChangedFile  `json:"changed_files,omitempty"`
}

// Sentinel errors for client-side branching (e.g. HTTP status mapping).
var (
    // ErrInvalidInput indicates malformed or missing input (bad URL, empty options, etc.).
    ErrInvalidInput = errors.New("invalid input")

    // ErrUnsupportedProvider indicates the provider is not recognized or plugin is not installed.
    ErrUnsupportedProvider = errors.New("unsupported provider")

    // ErrAuthRequired indicates authentication is missing or failed.
    ErrAuthRequired = errors.New("authentication required")

    // ErrUpstreamFailure indicates the upstream provider API failed (network, 5xx, timeout).
    ErrUpstreamFailure = errors.New("upstream failure")
)

// Analyze fetches a pull request and returns structured analysis.
func Analyze(ctx context.Context, opts AnalyzeOptions) (Result, error)

// Prompt fetches a pull request and returns a review prompt string.
func Prompt(ctx context.Context, opts AnalyzeOptions) (string, error)
```

### What remains internal

- `internal/domain` — domain models (used by Result via mapping, not re-exported)
- `internal/provider` — provider implementations and registry
- `internal/classifier` — change type classification
- `internal/analyzer` — risk estimation, review axes
- `internal/formatter` — JSON/Markdown/text rendering
- `internal/prompt` — prompt rendering and templates
- `internal/usecase` — use case orchestration
- `internal/fileutil` — file heuristics
- `internal/config` — CLI configuration loading

These MUST NOT be imported outside this module.

---

## Rationale

### 1. Why `internal/` alone is not enough

Go's `internal/` rule prevents any external module from importing these packages. Without a public surface, prism cannot be used as a library by `prism-api`, editor plugins, or any other consumer. The design goal of v0.3.0+ requires library usage, so this must change.

### 2. Why a single high-level package (`pkg/prism`)

Exposing `pkg/domain`, `pkg/usecase`, `pkg/analyzer`, etc. individually would lock prism's internal design. Every rename, split, or merge of internal packages would become a breaking change. Maintaining backwards compatibility across dozens of exported types is a long-term maintenance tax.

A single high-level package with a minimal function surface (`Analyze`, `Prompt`) gives us:

- **Freedom to refactor internals** — the public API is a stable facade
- **Clarity for consumers** — one obvious entry point, no decision paralysis
- **Symmetry** — CLI and API both call `pkg/prism` the same way, so behavior cannot diverge

The `AnalyzeOptions` struct is designed to be **additive-only**: new optional fields can be added without breaking existing callers.

### 3. What is NOT exposed (and why)

- **Domain models** (`PullRequest`, `ChangedFile`, etc.) — these change as the analysis pipeline evolves. Exposing them would freeze the shape of PR data.
- **Provider interfaces** — third parties cannot implement custom providers in-process. They must use the external plugin binary protocol (documented in [provider-plugin-protocol.md](../provider-plugin-protocol.md)).
- **Classifier / analyzer internals** — these are implementation details of the compiler, not a public library.
- **Formatter functions** — callers work with `Result` as structured data, not serialized strings. If string output is needed, callers format `Result` themselves.
- **Prompt templates** — templates are internal. Custom templates are still supported via `--template` at the CLI level, which reads from a file path.

### 4. Differences from the CLI JSON schema

`Result` is **structurally similar** to the CLI JSON output (`docs/json-schema.md`) but not identical. The differences are intentional:

| Field | CLI JSON | `pkg/prism.Result` | Reason |
|-------|----------|---------------------|--------|
| `provider` | top-level | `pull_request.provider` | Library consumers treat the PR as a self-contained object; provider belongs with PR metadata |
| `pull_request.description` | included | **excluded** | Descriptions can be large and are rarely needed by programmatic consumers; reduces response size |
| `pull_request.url` | not present | **included** | Avoids requiring consumers to reconstruct the URL from owner/repo/id |

These differences will be unified in **Phase 2** when the CLI is refactored to use `pkg/prism` internally. At that point, the CLI JSON schema will align with `pkg/prism.Result` (a potential breaking change to CLI output, to be announced in release notes).

Until then, consumers who need the exact CLI JSON shape should use the CLI binary; consumers who want the library API should use `pkg/prism`.

### 5. Compatibility guarantees

Within a **major version** (v1, v2, ...), `pkg/prism` guarantees:

- **Function signatures** of `Analyze` and `Prompt` do not change
- **Existing fields** on `AnalyzeOptions` and `Result` do not change type or meaning
- **New fields** may be added to `AnalyzeOptions` (as optional) and `Result` (as additive)
- **Default behavior** does not change in backwards-incompatible ways

Breaking changes require a new major version (`v2`, `v3`).

Before v1.0.0, the public API is considered unstable and may change in minor versions, but changes will be minimized and documented in release notes.

### 6. How CLI, API, and future clients use this

**CLI** (`cmd/prism`):
- Parses flags, loads config, builds `AnalyzeOptions`
- Calls `prism.Analyze()` / `prism.Prompt()`
- Formats `Result` into JSON/Markdown/text for output

**prism-api** (separate repo):
- HTTP handler parses request, builds `AnalyzeOptions`
- Calls `prism.Analyze()` / `prism.Prompt()`
- Serializes `Result` to JSON response

**Third-party clients** (editor plugins, CI tools):
- Import `github.com/hidetzu/prism/pkg/prism`
- Call the same functions
- Receive the same `Result` type

All clients use the same entry point, ensuring behavioral consistency.

---

## Consequences

### Positive
- CLI and API stay in sync because both use `pkg/prism`
- Internal refactoring is possible without breaking consumers
- Clear public surface for documentation and testing
- Foundation for prism-api and future library clients
- Plugin architecture remains the only extension point for providers (reinforcing the provider plugin protocol)

### Negative
- Some code must move from `internal/usecase` to `pkg/prism` (or `pkg/prism` must import `internal/usecase`, which is allowed)
- `Result` struct must be carefully designed — its shape is now part of the public contract
- Compatibility discipline required: changes to `pkg/prism` must be reviewed with backwards-compatibility in mind
- Custom formatters / custom providers remain impossible in-process (by design — plugins only)

---

## Alternatives Considered

### A. Expose internal packages as public (`pkg/domain`, `pkg/usecase`, ...)
Rejected. Freezes internal design across all public packages. High maintenance tax.

### B. Keep everything internal, build prism-api inside the prism module
Rejected. Couples API deployment to prism release cycle. Prevents third-party library usage. Violates separation of concerns.

### C. Expose only types (`pkg/domain`), keep logic internal
Rejected. Third parties could read domain types but not run analysis. Half-solution that still freezes domain shape.

### D. Single `pkg/prism` high-level facade (this ADR)
Accepted. Minimal public surface, maximum internal freedom, clean consumer story.

---

## Implementation Plan

### Phase 1 (this PR)

1. Create `pkg/prism/prism.go` with `AnalyzeOptions`, `Result`, `Analyze`, `Prompt`, and sentinel errors
2. Implement `Analyze` / `Prompt` by composing `internal/classifier`, `internal/analyzer`, `internal/provider`, `internal/prompt`
3. Add unit tests for `pkg/prism` covering validation and error categorization
4. Document the public API in `docs/public-api.md` or README

### Phase 2 (subsequent PR)

5. Refactor `cmd/prism` to use `pkg/prism` internally
   - `analyze` JSON format: call `pkg/prism.Analyze()` and `json.Marshal(result)`
   - `analyze` markdown/text format: either reuse existing `internal/formatter` (taking `Result`) or keep current path
   - `prompt` command: switch to `pkg/prism.Prompt()`
6. Verify CLI JSON output matches `pkg/prism.Result` JSON exactly via a golden test

### Phase 3 (after CLI refactor)

7. Create `prism-api` repository
8. Tag prism v0.3.0 once the public API is validated

---

## Follow-up Work

- Implement `pkg/prism` following the shape defined in this ADR
- Refactor `cmd/prism` to call `pkg/prism` instead of `internal/usecase`
- Create prism-api repository (depends on this ADR being implemented)
- Document the public API in README with usage examples
