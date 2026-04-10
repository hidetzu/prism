# Versioning Policy

prism follows [Semantic Versioning 2.0.0](https://semver.org/).

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

## Pre-1.0

During v0.x development, the JSON schema may change between minor versions. Breaking changes will be documented in release notes.

## Golden Tests

JSON output is verified by golden tests in `testdata/`. If golden test files need updating due to output changes, the change must be reviewed explicitly to ensure backward compatibility is intentional.
