# ADR-0001: Provider Plugin Architecture for v0.2.0

## Status
Accepted

## Date
2026-04-11

## Context

prism v0.2.0 aims to add AWS CodeCommit Pull Request support.
However, prism's core responsibility is **decomposing Pull Requests into structured review context**, not absorbing API implementations and authentication mechanisms for every hosting service.

Two concerns drove this decision:

1. **Provider detection design**
   - The current implementation assumes GitHub
   - Needs a strategy for extending to GitHub / CodeCommit / future Bitbucket, etc.

2. **Authentication and dependency differences**
   - GitHub uses `GITHUB_TOKEN` with Bearer authentication
   - CodeCommit requires AWS credential chain / IAM / region
   - Embedding the AWS SDK directly into prism would bloat dependencies and blur responsibilities

Additionally, an existing tool `ccpr` already handles CodeCommit PR retrieval. This asset should be leveraged while preserving prism's core responsibilities.

---

## Decision

v0.2.0 extends providers via **external binary plugins**.

prism itself holds only the provider interface and plugin execution layer. CodeCommit support is implemented as a separate binary that wraps `ccpr`.

### Approach

- Provider extensibility uses the **external binary approach**
- Plugins are executable binaries such as `prism-provider-codecommit`
- prism invokes plugins as subprocesses
- Plugins return PR data in a standardized JSON format
- prism converts the returned JSON into `PullRequest` domain models and continues analysis

### GitHub provider handling

- In v0.2.0, **GitHub remains built-in**
- This preserves the user experience of working with GitHub immediately after `go install`
- Internally, the boundary is kept equivalent to plugins, allowing future extraction
- CodeCommit and subsequent providers use the plugin approach

### Provider detection strategy

- Default: **auto-detect from URL**
- When `--provider` is explicitly specified, **URL-based auto-detection is skipped**
- `--provider` enables use with GitHub Enterprise and other hosts where URL auto-detection is insufficient

### Initial rules

- URLs containing `github.com` use the GitHub provider
- URLs matching CodeCommit patterns use the CodeCommit provider
- `--provider github` etc. forces the specified provider
- When auto-detection fails, a clear error message suggests using `--provider`

---

## Rationale

### 1. Preserves prism's responsibility
prism's essence is not fetching PRs, but **transforming fetched PRs into review context**. Absorbing provider-specific implementations would blur this responsibility.

### 2. Leverages ccpr
For CodeCommit support, the existing `ccpr` tool can be reused. By wrapping it as an external provider rather than embedding the AWS SDK into prism, asset reuse and responsibility separation are achieved simultaneously.

### 3. Strong future extensibility
When considering future providers like Bitbucket, GitHub Enterprise, and GitLab, the plugin approach minimizes changes to prism itself.

### 4. Clear dependency separation
The GitHub provider remains lightweight, while the CodeCommit provider can independently manage AWS authentication and SDK dependencies. This keeps the core build, maintenance, and distribution simple.

---

## Alternatives Considered

## A. External binary approach
Adopted.

### Advantages
- Complete dependency isolation
- Clear prism core responsibility
- Easy to leverage ccpr
- Easy to add future providers

### Disadvantages
- Plugin input/output specification must be stabilized
- Subprocess execution and error handling required
- Distribution method must be designed

## B. Go internal approach
Rejected.

### Reasons
- Provider-specific dependencies would leak into prism core
- AWS SDK and similar dependencies would mix into the core
- Violates prism's principle of responsibility separation

---

## Consequences

### Positive
- CodeCommit support added in v0.2.0 while keeping the core simple
- Path opened for reusing ccpr as a provider plugin
- Future provider extension strategy made explicit

### Negative
- Plugin protocol must be designed
- Plugin discovery specification required
- User-facing installation instructions become slightly more complex

---

## Plugin Protocol

See [Provider Plugin Protocol](../provider-plugin-protocol.md) for invocation conventions, JSON schema, and discovery specification.

---

## Implementation Status

- [x] Define plugin protocol JSON schema
- [x] Implement provider registry
- [x] Implement provider auto-detection logic
- [x] Implement `--provider` override rule
- [x] Adopt ccpr wrapper approach for CodeCommit plugin → [prism-provider-codecommit](https://github.com/hidetzu/prism-provider-codecommit)
