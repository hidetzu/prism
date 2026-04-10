# Use Cases

## UC-1: Analyze a GitHub PR as JSON

**Actor:** Developer or CI pipeline

**Input:**
```bash
prism analyze https://github.com/owner/repo/pull/123 --format json
```

**Flow:**
1. Parse PR URL to extract provider, owner, repo, PR number
2. Fetch PR metadata, changed files, and diffs via GitHub API
3. Classify change type from PR title, description, and file paths
4. Estimate risk level based on change scope and affected areas
5. Suggest review axes based on classification and file characteristics
6. Output structured JSON to stdout

**Output:** JSON containing PR metadata, change type, risk level, review axes, related files, and warnings.

---

## UC-2: Generate a light review prompt

**Actor:** Developer piping output to AI

**Input:**
```bash
prism prompt https://github.com/owner/repo/pull/123 --mode light
```

**Flow:**
1. Run analysis pipeline (same as UC-1)
2. Select light template (minimal context, top review axes)
3. Render prompt text

**Output:** A focused prompt suitable for quick AI screening.

---

## UC-3: Generate a detailed review prompt

**Actor:** Developer seeking thorough AI review

**Input:**
```bash
prism prompt https://github.com/owner/repo/pull/123 --mode detailed
```

**Flow:**
1. Run analysis pipeline
2. Select detailed template (full context, expanded review axes, patch excerpts)
3. Render prompt text

**Output:** A comprehensive prompt with full context for deep AI review.

---

## UC-4: Generate a cross-file review prompt

**Actor:** Developer reviewing architectural changes

**Input:**
```bash
prism prompt https://github.com/owner/repo/pull/123 --mode cross
```

**Flow:**
1. Run analysis pipeline
2. Select cross template (module structure, interface/config/test relationships)
3. Render prompt text

**Output:** A prompt focused on cross-file consistency and integration.

---

## UC-5: Pipe analysis to Claude for review

**Actor:** Developer using Claude Code

**Input:**
```bash
prism analyze https://github.com/owner/repo/pull/123 --format json | claude -p "Review this pull request"
```

**Flow:**
1. prism outputs structured JSON to stdout
2. Claude Code receives JSON as input context
3. Claude generates review based on structured context

**Output:** AI review with consistent quality thanks to standardized input.

---

## UC-6: Debug PR data fetching

**Actor:** Developer troubleshooting prism behavior

**Input:**
```bash
prism fetch https://github.com/owner/repo/pull/123 --format json
```

**Flow:**
1. Parse PR URL
2. Fetch raw PR data from provider
3. Output without analysis

**Output:** Raw PR data for inspection.
