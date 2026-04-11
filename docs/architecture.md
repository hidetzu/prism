# Architecture

## Overview

prism follows a layered architecture with clear separation between external data sources (providers), domain logic, and output formatting.

```mermaid
graph TD
    CLI["CLI<br/><code>cmd/prism</code>"]
    UC["Use Cases<br/><code>internal/usecase</code>"]
    PRV["Providers<br/><code>internal/provider</code>"]
    CLS["Classifier<br/><code>internal/classifier</code>"]
    ANL["Analyzer<br/><code>internal/analyzer</code>"]
    DOM["Domain Models<br/><code>internal/domain</code>"]
    FMT["Formatter<br/><code>internal/formatter</code>"]
    PMT["Prompt Renderer<br/><code>internal/prompt</code>"]

    CLI --> UC
    UC --> PRV
    UC --> CLS
    UC --> ANL
    UC --> FMT
    UC --> PMT

    PRV --> DOM
    CLS --> DOM
    ANL --> DOM
    FMT --> DOM
    PMT --> DOM

    PRV --- REG["Registry"]
    REG --- GH["GitHub<br/><i>built-in</i>"]
    REG --- PLG["Plugin Executor"]
    PLG --- CC["prism-provider-codecommit"]
    PLG --- ETC["prism-provider-..."]

    FMT --- JSON["JSON"]
    FMT --- MD["Markdown"]
    FMT --- TXT["Text"]

    PMT --- LT["light"]
    PMT --- DT["detailed"]
    PMT --- CR["cross"]

    style DOM fill:#e8f4f8,stroke:#2196f3
    style CLI fill:#f3e8f4,stroke:#9c27b0
    style UC fill:#f4f0e8,stroke:#ff9800
```

### Data Flow

```mermaid
flowchart LR
    A[PR URL] --> B[Provider<br/>fetch]
    B --> C[PullRequest<br/>domain model]
    C --> D[Classifier]
    D --> E[Analyzer]
    E --> F[AnalysisResult]
    F --> G{Output}
    G -->|analyze| H[Formatter<br/>JSON / Markdown]
    G -->|prompt| I[Prompt Renderer<br/>light / detailed / cross]
```

## Package Responsibilities

### `pkg/prism`

Public API for embedding prism as a library. The only exported package outside of `cmd/`.

Exposes `Analyze` and `Prompt` functions, stable input/output types (`AnalyzeOptions`, `Result`, `PRInfo`, `AnalysisResult`, `ChangedFile`), and sentinel errors (`ErrInvalidInput`, `ErrUnsupportedProvider`, `ErrAuthRequired`, `ErrUpstreamFailure`) for client-side branching.

Consumers:

- **cmd/prism** — the CLI (refactor to use pkg/prism is Phase 2)
- **prism-api** — HTTP service (planned)
- **Editor / IDE plugins** — library consumers
- **CI / automation tools** — library consumers

See [ADR-0002](adr/0002-public-api-boundary.md) for the design rationale and compatibility policy.

### `cmd/prism`

CLI entrypoint. Parses arguments, resolves configuration, and delegates to use cases. Should remain thin.

### `internal/domain`

Core domain models shared across the entire application. No external dependencies.

- `PullRequest` — PR metadata, changed files, diff summary
- `ChangedFile` — per-file change details (path, status, additions, deletions, patch, language flags)
- `AnalysisResult` — classification output (change type, risk, review axes, warnings)
- `PromptBundle` — assembled prompt (mode, system prompt, user prompt, attached context)
- `PRRef` — provider-agnostic PR reference (provider, owner, repo, PR number)

### `internal/provider`

Adapters for external PR sources. Each provider implements the `Provider` interface:

```go
type Provider interface {
    Parse(input string) (PRRef, error)
    FetchPullRequest(ctx context.Context, ref PRRef) (PullRequest, error)
}
```

All provider-specific data is normalized into domain models at the provider boundary.

The provider layer consists of:

- **Registry** — resolves a provider by name or auto-detects from URL. GitHub is built-in; other providers are discovered as external plugin binaries (`prism-provider-<name>`) on PATH.
- **Plugin Executor** — runs an external provider binary as a subprocess, parses its stdout JSON into domain models, and enriches files with language/test/config classification.
- **Plugin Protocol** — plugins are invoked as `prism-provider-<name> fetch <PR_URL>` and return a JSON object to stdout. See [ADR-0001](adr/0001-provider-plugin-architecture.md) for details.

### `internal/classifier`

Determines change type based on PR title, description, file paths, and diff content.

Output categories: `feature`, `bugfix`, `refactor`, `test-only`, `docs-only`, `config-change`, `dependency-update`, `infra-change`.

### `internal/analyzer`

Estimates risk level and suggests review axes based on classification results and file characteristics.

### `internal/formatter`

Serializes `AnalysisResult` into output formats (JSON, Markdown, text).

### `internal/prompt`

Renders `PromptBundle` for each review mode using templates.

### `internal/usecase`

Orchestrates the pipeline: fetch → classify → analyze → format/render. Each use case corresponds to a CLI command.

## Design Principles

1. **Provider abstraction first** — New PR sources are added as external plugin binaries without modifying prism itself
2. **Domain models are the contract** — All packages communicate through domain types
3. **Output stability** — JSON schema must remain backward-compatible within a major version
4. **Testability** — All external dependencies are behind interfaces; use fixtures and golden files for output verification
5. **No LLM dependency** — prism compiles context, it does not invoke AI
