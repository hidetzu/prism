# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.3.0] - 2026-04-11

### Added
- Public library API at `pkg/prism` (`Analyze`, `Prompt`, sentinel errors)
- `pkg/prism` happy-path integration tests with mock GitHub server
- Test seam: `provider.NewRegistryWithGitHubBaseURL` for redirecting the GitHub provider to a local httptest server
- Foundation for `prism-api` (HTTP service, separate repository)
- ADR-0002: Public API Boundary

### Changed
- CLI `analyze` command now uses `pkg/prism.Analyze()` internally
- CLI JSON output is now byte-identical to `pkg/prism.Result` (verified by golden test)
- `internal/formatter` rewritten to take `prism.Result` as input
- CLI exit-code mapping recognizes `pkg/prism` sentinel errors (`ErrInvalidInput`, `ErrUnsupportedProvider`, `ErrAuthRequired`, `ErrUpstreamFailure`)
- README, `docs/architecture.md`, `docs/json-schema.md` updated to reflect the unified pipeline

### Removed
- `internal/usecase/analyze.go` (replaced by `pkg/prism.Analyze` + `internal/formatter`)
- `internal/formatter/types.go` (formatter now uses `prism.Result` directly)

### Breaking
CLI JSON output structure changed to align with `pkg/prism.Result`:

- `provider` moved from top-level to `pull_request.provider`
- `pull_request.description` removed (use `prism fetch` for raw description)
- `pull_request.url` added (canonical PR URL)
- `changed_files[].patch` no longer included by default (library has `IncludePatches` opt-in)
- Empty/zero fields are now omitted (`omitempty` semantics)

These changes were made early in the project lifecycle. Impact is expected to be minimal. See [docs/json-schema.md](docs/json-schema.md#breaking-changes-in-v030) for the field-by-field migration guide.

`prism prompt` and `prism fetch` commands are unchanged in this release; they still go through `internal/usecase` pending future extension of `pkg/prism.Prompt`.

## [v0.2.0] - 2026-04-11

### Added
- Provider plugin architecture: external `prism-provider-<name>` binaries discovered on PATH
- `--provider` flag for explicit provider selection (auto-detected from URL if omitted)
- Provider plugin protocol v1 (versioned JSON contract for plugin authors)
- AWS CodeCommit support via [prism-provider-codecommit](https://github.com/hidetzu/prism-provider-codecommit) plugin
- ADR-0001: Provider Plugin Architecture
- HTTP request timeout (30s) and pagination upper bound on the GitHub provider

## [v0.1.0] - 2026-04-11

### Added
- Initial release: GitHub provider, `analyze`/`prompt`/`fetch` commands
- Output formats: JSON, Markdown, text
- Prompt modes: light, detailed, cross
- Language support: English, Japanese
- Custom template support (`--template`)
- Configuration file (`~/.config/prism/config.yaml`)
- Exit codes 0–4 for CI/CD integration

[Unreleased]: https://github.com/hidetzu/prism/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/hidetzu/prism/releases/tag/v0.3.0
[v0.2.0]: https://github.com/hidetzu/prism/releases/tag/v0.2.0
[v0.1.0]: https://github.com/hidetzu/prism/releases/tag/v0.1.0
